package handlers

import (
	"context"
	"errors"

	"github.com/boxingoctopus/snackmates/api/internal/friends"
	"github.com/boxingoctopus/snackmates/api/internal/middleware"
	"github.com/boxingoctopus/snackmates/api/internal/models"
	"github.com/boxingoctopus/snackmates/api/internal/notifications"
	"github.com/boxingoctopus/snackmates/api/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FriendHandler struct {
	pool    *pgxpool.Pool
	storage *storage.Client
	hub     *notifications.Hub
}

func NewFriendHandler(pool *pgxpool.Pool, s *storage.Client, hub *notifications.Hub) *FriendHandler {
	return &FriendHandler{pool: pool, storage: s, hub: hub}
}

func (h *FriendHandler) RegisterRoutes(app fiber.Router) {
	f := app.Group("/friends", middleware.RequireAuth(h.pool))
	f.Get("/", h.List)
	f.Get("/requests", h.ListRequests)
	f.Post("/request", h.Request)
	f.Post("/:id/accept", h.Accept)
	f.Post("/:id/decline", h.Decline)
	f.Delete("/:id", h.Remove)
}

func (h *FriendHandler) List(c *fiber.Ctx) error {
	list, err := friends.ListAccepted(c.Context(), h.pool, middleware.GetUserID(c))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	enrichFriendUsers(c.Context(), h.pool, h.storage, list)
	return c.JSON(list)
}

func (h *FriendHandler) ListRequests(c *fiber.Ctx) error {
	list, err := friends.ListIncoming(c.Context(), h.pool, middleware.GetUserID(c))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	enrichFriendUsers(c.Context(), h.pool, h.storage, list)
	return c.JSON(list)
}

func (h *FriendHandler) Request(c *fiber.Ctx) error {
	var body struct {
		Username string `json:"username"`
	}
	if err := c.BodyParser(&body); err != nil || body.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username required"})
	}

	rec, err := friends.Request(c.Context(), h.pool, middleware.GetUserID(c), body.Username)
	if err != nil {
		return friendError(c, err)
	}
	h.hub.Notify(rec.AddresseeID)
	return c.Status(fiber.StatusCreated).JSON(toFriendshipResponse(*rec, nil))
}

func (h *FriendHandler) Accept(c *fiber.Ctx) error {
	friendshipID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid friendship id"})
	}

	rec, err := friends.Accept(c.Context(), h.pool, friendshipID, middleware.GetUserID(c))
	if err != nil {
		return friendError(c, err)
	}
	h.hub.Notify(rec.RequesterID, rec.AddresseeID)
	return c.JSON(toFriendshipResponse(*rec, nil))
}

func (h *FriendHandler) Decline(c *fiber.Ctx) error {
	friendshipID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid friendship id"})
	}

	rec, err := friends.Decline(c.Context(), h.pool, friendshipID, middleware.GetUserID(c))
	if err != nil {
		return friendError(c, err)
	}
	h.hub.Notify(rec.RequesterID, rec.AddresseeID)
	return c.JSON(fiber.Map{"ok": true})
}

func (h *FriendHandler) Remove(c *fiber.Ctx) error {
	friendshipID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid friendship id"})
	}

	rec, err := friends.Remove(c.Context(), h.pool, friendshipID, middleware.GetUserID(c))
	if err != nil {
		return friendError(c, err)
	}
	h.hub.Notify(rec.RequesterID, rec.AddresseeID)
	return c.JSON(fiber.Map{"ok": true})
}

func friendError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, friends.ErrUserNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, friends.ErrSelfFriend),
		errors.Is(err, friends.ErrAlreadyFriends),
		errors.Is(err, friends.ErrRequestPending),
		errors.Is(err, friends.ErrIncomingRequest):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, friends.ErrFriendshipNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, friends.ErrNotAuthorized):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
}

func toFriendshipResponse(rec friends.Record, user *models.PublicUser) models.Friendship {
	return models.Friendship{
		ID:          rec.ID,
		RequesterID: rec.RequesterID,
		AddresseeID: rec.AddresseeID,
		Status:      rec.Status,
		CreatedAt:   rec.CreatedAt,
		UpdatedAt:   rec.UpdatedAt,
		User:        user,
	}
}

func enrichFriendUsers(ctx context.Context, pool *pgxpool.Pool, s *storage.Client, list []models.Friendship) {
	for i := range list {
		if list[i].User == nil {
			continue
		}
		var avatarKey, avatarURL *string
		_ = pool.QueryRow(ctx, `
			SELECT avatar_key, avatar_url FROM users WHERE id = $1
		`, list[i].User.ID).Scan(&avatarKey, &avatarURL)
		if s == nil {
			if avatarURL != nil && *avatarURL != "" {
				list[i].User.AvatarURL = avatarURL
			}
			continue
		}
		resolved, err := s.ResolveObjectURL(ctx, avatarURL, avatarKey)
		if err == nil && resolved != "" {
			list[i].User.AvatarURL = &resolved
		}
	}
}
