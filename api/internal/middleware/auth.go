package middleware

import (
	"strings"

	"github.com/boxingoctopus/snackmates/api/internal/auth"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const UserIDKey = "userID"

func RequireAuth(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractToken(c)
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "authentication required"})
		}
		userID, err := auth.ValidateSession(c.Context(), pool, token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
		}
		c.Locals(UserIDKey, userID)
		return c.Next()
	}
}

func OptionalAuth(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractToken(c)
		if token != "" {
			if userID, err := auth.ValidateSession(c.Context(), pool, token); err == nil {
				c.Locals(UserIDKey, userID)
			}
		}
		return c.Next()
	}
}

func extractToken(c *fiber.Ctx) string {
	authHeader := c.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	if token := c.Query("access_token"); token != "" {
		return token
	}
	return c.Cookies("session")
}

func GetUserID(c *fiber.Ctx) uuid.UUID {
	if v, ok := c.Locals(UserIDKey).(uuid.UUID); ok {
		return v
	}
	return uuid.Nil
}
