package handlers

import (
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/gofiber/fiber/v2"

	"github.com/boxingoctopus/snackmates/api/internal/auth"
	"github.com/boxingoctopus/snackmates/api/internal/middleware"
)

func (h *AuthHandler) WebAuthnRegisterFinishRaw(c *fiber.Ctx) error {
	sessionData := c.Get("X-WebAuthn-Session")
	if sessionData == "" {
		var body struct {
			SessionData string `json:"session_data"`
		}
		_ = c.BodyParser(&body)
		sessionData = body.SessionData
	}
	parsed, err := protocol.ParseCredentialCreationResponseBytes(c.Body())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	deviceName := c.Get("X-Device-Name")
	if err := h.webauthn.FinishRegistration(c.Context(), h.pool, middleware.GetUserID(c), sessionData, parsed, deviceName); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *AuthHandler) WebAuthnLoginFinishRaw(c *fiber.Ctx) error {
	sessionData := c.Get("X-WebAuthn-Session")
	if sessionData == "" {
		var body struct {
			SessionData string `json:"session_data"`
		}
		_ = c.BodyParser(&body)
		sessionData = body.SessionData
	}
	parsed, err := protocol.ParseCredentialRequestResponseBytes(c.Body())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	userID := middleware.GetUserID(c)
	if err := h.webauthn.FinishLogin(c.Context(), h.pool, userID, sessionData, parsed); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	token, err := auth.CreateSession(c.Context(), h.pool, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	setSessionCookie(c, token)
	return c.JSON(fiber.Map{"token": token, "ok": true})
}
