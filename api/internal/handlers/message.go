package handlers

import (
	"context"
	"errors"

	"github.com/boxingoctopus/snackmates/api/internal/messages"
	"github.com/boxingoctopus/snackmates/api/internal/middleware"
	"github.com/boxingoctopus/snackmates/api/internal/models"
	"github.com/boxingoctopus/snackmates/api/internal/notifications"
	"github.com/boxingoctopus/snackmates/api/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageHandler struct {
	pool    *pgxpool.Pool
	storage *storage.Client
	hub     *notifications.Hub
}

func NewMessageHandler(pool *pgxpool.Pool, s *storage.Client, hub *notifications.Hub) *MessageHandler {
	return &MessageHandler{pool: pool, storage: s, hub: hub}
}

func (h *MessageHandler) RegisterRoutes(app fiber.Router) {
	m := app.Group("/messages", middleware.RequireAuth(h.pool))
	m.Get("/conversations", h.ListConversations)
	m.Post("/conversations", h.StartConversation)
	m.Get("/conversations/:id", h.GetConversation)
	m.Post("/conversations/:id/messages", h.SendMessage)
	m.Post("/conversations/:id/read", h.MarkRead)
}

func (h *MessageHandler) ListConversations(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	list, err := messages.ListConversations(c.Context(), h.pool, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	enrichConversationUsers(c.Context(), h.pool, h.storage, list)

	unread, err := messages.UnreadTotal(c.Context(), h.pool, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"unread_count":  unread,
		"conversations": list,
	})
}

func (h *MessageHandler) StartConversation(c *fiber.Ctx) error {
	var body struct {
		Username string `json:"username"`
	}
	if err := c.BodyParser(&body); err != nil || body.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username required"})
	}

	rec, err := messages.FindOrCreate(c.Context(), h.pool, middleware.GetUserID(c), body.Username)
	if err != nil {
		return messageError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": rec.ID})
}

func (h *MessageHandler) GetConversation(c *fiber.Ctx) error {
	conversationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid conversation id"})
	}

	userID := middleware.GetUserID(c)
	conv, err := messages.GetByID(c.Context(), h.pool, conversationID)
	if err != nil {
		return messageError(c, err)
	}
	if conv.UserAID != userID && conv.UserBID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": messages.ErrNotAuthorized.Error()})
	}

	msgs, err := messages.ListMessages(c.Context(), h.pool, conversationID, userID, 100)
	if err != nil {
		return messageError(c, err)
	}

	_ = messages.MarkRead(c.Context(), h.pool, conversationID, userID)

	otherUser, err := messages.OtherUser(c.Context(), h.pool, *conv, userID)
	if err != nil {
		return messageError(c, err)
	}
	enrichPublicUser(c.Context(), h.pool, h.storage, otherUser)

	resp := fiber.Map{
		"conversation": fiber.Map{
			"id":         conv.ID,
			"user_a_id":  conv.UserAID,
			"user_b_id":  conv.UserBID,
			"created_at": conv.CreatedAt,
			"updated_at": conv.UpdatedAt,
			"other_user": otherUser,
		},
		"messages": msgs,
	}
	return c.JSON(resp)
}

func (h *MessageHandler) SendMessage(c *fiber.Ctx) error {
	conversationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid conversation id"})
	}

	var body struct {
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	senderID := middleware.GetUserID(c)
	msg, conv, err := messages.Send(c.Context(), h.pool, conversationID, senderID, body.Subject, body.Body)
	if err != nil {
		return messageError(c, err)
	}

	recipientID := conv.UserAID
	if recipientID == senderID {
		recipientID = conv.UserBID
	}
	h.hub.Notify(senderID, recipientID)

	return c.Status(fiber.StatusCreated).JSON(msg)
}

func (h *MessageHandler) MarkRead(c *fiber.Ctx) error {
	conversationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid conversation id"})
	}

	if err := messages.MarkRead(c.Context(), h.pool, conversationID, middleware.GetUserID(c)); err != nil {
		return messageError(c, err)
	}
	return c.JSON(fiber.Map{"ok": true})
}

func messageError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, messages.ErrUserNotFound),
		errors.Is(err, messages.ErrConversationNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, messages.ErrNotAuthorized):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, messages.ErrCannotMessage):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, messages.ErrEmptyBody),
		errors.Is(err, messages.ErrEmptySubject),
		errors.Is(err, messages.ErrBodyTooLong),
		errors.Is(err, messages.ErrSubjectTooLong):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
}

func enrichConversationUsers(ctx context.Context, pool *pgxpool.Pool, s *storage.Client, list []models.Conversation) {
	for i := range list {
		if list[i].OtherUser != nil {
			enrichPublicUser(ctx, pool, s, list[i].OtherUser)
		}
	}
}

func enrichPublicUser(ctx context.Context, pool *pgxpool.Pool, s *storage.Client, user *models.PublicUser) {
	if user == nil {
		return
	}
	var avatarKey, avatarURL *string
	_ = pool.QueryRow(ctx, `
		SELECT avatar_key, avatar_url FROM users WHERE id = $1
	`, user.ID).Scan(&avatarKey, &avatarURL)
	if s == nil {
		if avatarURL != nil && *avatarURL != "" {
			user.AvatarURL = avatarURL
		}
		return
	}
	resolved, err := s.ResolveObjectURL(ctx, avatarURL, avatarKey)
	if err == nil && resolved != "" {
		user.AvatarURL = &resolved
	}
}
