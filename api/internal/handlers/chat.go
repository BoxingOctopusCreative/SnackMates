package handlers

import (
	"context"
	"errors"

	"github.com/boxingoctopus/snackmates/api/internal/chats"
	"github.com/boxingoctopus/snackmates/api/internal/middleware"
	"github.com/boxingoctopus/snackmates/api/internal/models"
	"github.com/boxingoctopus/snackmates/api/internal/notifications"
	"github.com/boxingoctopus/snackmates/api/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatHandler struct {
	pool    *pgxpool.Pool
	storage *storage.Client
	hub     *notifications.Hub
}

func NewChatHandler(pool *pgxpool.Pool, s *storage.Client, hub *notifications.Hub) *ChatHandler {
	return &ChatHandler{pool: pool, storage: s, hub: hub}
}

func (h *ChatHandler) RegisterRoutes(app fiber.Router) {
	c := app.Group("/chats", middleware.RequireAuth(h.pool))
	c.Get("/", h.ListChats)
	c.Post("/", h.StartChat)
	c.Get("/:id", h.GetChat)
	c.Post("/:id/messages", h.SendMessage)
	c.Post("/:id/read", h.MarkRead)
}

func (h *ChatHandler) ListChats(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	list, err := chats.ListChats(c.Context(), h.pool, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	enrichChatUsers(c.Context(), h.pool, h.storage, list)

	unread, err := chats.UnreadTotal(c.Context(), h.pool, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"unread_count": unread,
		"chats":        list,
	})
}

func (h *ChatHandler) StartChat(c *fiber.Ctx) error {
	var body struct {
		Username string `json:"username"`
	}
	if err := c.BodyParser(&body); err != nil || body.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username required"})
	}

	rec, err := chats.FindOrCreate(c.Context(), h.pool, middleware.GetUserID(c), body.Username)
	if err != nil {
		return chatError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": rec.ID})
}

func (h *ChatHandler) GetChat(c *fiber.Ctx) error {
	chatID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid chat id"})
	}

	userID := middleware.GetUserID(c)
	chat, err := chats.GetByID(c.Context(), h.pool, chatID)
	if err != nil {
		return chatError(c, err)
	}
	if chat.UserAID != userID && chat.UserBID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": chats.ErrNotAuthorized.Error()})
	}

	msgs, err := chats.ListMessages(c.Context(), h.pool, chatID, userID, 100)
	if err != nil {
		return chatError(c, err)
	}

	_ = chats.MarkRead(c.Context(), h.pool, chatID, userID)

	otherUser, err := chats.OtherUser(c.Context(), h.pool, *chat, userID)
	if err != nil {
		return chatError(c, err)
	}
	enrichPublicUser(c.Context(), h.pool, h.storage, otherUser)

	return c.JSON(fiber.Map{
		"chat": fiber.Map{
			"id":         chat.ID,
			"user_a_id":  chat.UserAID,
			"user_b_id":  chat.UserBID,
			"created_at": chat.CreatedAt,
			"updated_at": chat.UpdatedAt,
			"other_user": otherUser,
		},
		"messages": msgs,
	})
}

func (h *ChatHandler) SendMessage(c *fiber.Ctx) error {
	chatID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid chat id"})
	}

	var body struct {
		Body string `json:"body"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	senderID := middleware.GetUserID(c)
	msg, chat, err := chats.Send(c.Context(), h.pool, chatID, senderID, body.Body)
	if err != nil {
		return chatError(c, err)
	}

	recipientID := chat.UserAID
	if recipientID == senderID {
		recipientID = chat.UserBID
	}
	h.hub.Notify(senderID, recipientID)

	return c.Status(fiber.StatusCreated).JSON(msg)
}

func (h *ChatHandler) MarkRead(c *fiber.Ctx) error {
	chatID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid chat id"})
	}

	if err := chats.MarkRead(c.Context(), h.pool, chatID, middleware.GetUserID(c)); err != nil {
		return chatError(c, err)
	}
	return c.JSON(fiber.Map{"ok": true})
}

func chatError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, chats.ErrUserNotFound),
		errors.Is(err, chats.ErrChatNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, chats.ErrNotAuthorized):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, chats.ErrCannotChat):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, chats.ErrEmptyBody),
		errors.Is(err, chats.ErrBodyTooLong):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
}

func enrichChatUsers(ctx context.Context, pool *pgxpool.Pool, s *storage.Client, list []models.Chat) {
	for i := range list {
		if list[i].OtherUser != nil {
			enrichPublicUser(ctx, pool, s, list[i].OtherUser)
		}
	}
}
