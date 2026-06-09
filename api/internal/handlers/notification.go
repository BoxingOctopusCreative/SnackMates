package handlers

import (
	"bufio"
	"errors"
	"fmt"
	"time"

	"github.com/boxingoctopus/snackmates/api/internal/friends"
	"github.com/boxingoctopus/snackmates/api/internal/middleware"
	"github.com/boxingoctopus/snackmates/api/internal/models"
	"github.com/boxingoctopus/snackmates/api/internal/notifications"
	"github.com/boxingoctopus/snackmates/api/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationHandler struct {
	pool    *pgxpool.Pool
	storage *storage.Client
	hub     *notifications.Hub
}

func NewNotificationHandler(pool *pgxpool.Pool, s *storage.Client, hub *notifications.Hub) *NotificationHandler {
	return &NotificationHandler{pool: pool, storage: s, hub: hub}
}

func (h *NotificationHandler) RegisterRoutes(app fiber.Router) {
	app.Get("/notifications", middleware.RequireAuth(h.pool), h.List)
	app.Get("/notifications/stream", middleware.RequireAuth(h.pool), h.Stream)
	app.Post("/notifications/:id/acknowledge", middleware.RequireAuth(h.pool), h.Acknowledge)
}

type notificationItem struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	CreatedAt  string            `json:"created_at"`
	Friendship models.Friendship `json:"friendship"`
}

func (h *NotificationHandler) List(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	ctx := c.Context()

	requests, err := friends.ListIncoming(ctx, h.pool, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	enrichFriendUsers(ctx, h.pool, h.storage, requests)

	accepted, err := friends.ListAcceptedUnseenForRequester(ctx, h.pool, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	enrichFriendUsers(ctx, h.pool, h.storage, accepted)

	items := make([]notificationItem, 0, len(requests)+len(accepted))
	for _, req := range requests {
		items = append(items, notificationItem{
			ID:         req.ID.String(),
			Type:       "snack_mate_request",
			CreatedAt:  req.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Friendship: req,
		})
	}
	for _, req := range accepted {
		items = append(items, notificationItem{
			ID:         req.ID.String(),
			Type:       "snack_mate_accepted",
			CreatedAt:  req.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Friendship: req,
		})
	}

	return c.JSON(fiber.Map{
		"unread_count": len(items),
		"items":        items,
	})
}

func (h *NotificationHandler) Acknowledge(c *fiber.Ctx) error {
	friendshipID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid notification id"})
	}

	userID := middleware.GetUserID(c)
	if err := friends.AcknowledgeAcceptance(c.Context(), h.pool, friendshipID, userID); err != nil {
		if errors.Is(err, friends.ErrFriendshipNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	h.hub.Notify(userID)
	return c.JSON(fiber.Map{"ok": true})
}

func (h *NotificationHandler) Stream(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	events, unsubscribe := h.hub.Subscribe(userID)
	defer unsubscribe()

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		fmt.Fprintf(w, "event: connected\ndata: {}\n\n")
		if err := w.Flush(); err != nil {
			return
		}

		heartbeat := time.NewTicker(15 * time.Second)
		defer heartbeat.Stop()

		for {
			select {
			case _, ok := <-events:
				if !ok {
					return
				}
				fmt.Fprintf(w, "event: refresh\ndata: {}\n\n")
				if err := w.Flush(); err != nil {
					return
				}
			case <-heartbeat.C:
				fmt.Fprintf(w, ": keepalive\n\n")
				if err := w.Flush(); err != nil {
					return
				}
			}
		}
	})

	return nil
}
