package handlers

import (
	"errors"
	"net/url"
	"time"

	"github.com/boxingoctopus/snackmates/api/internal/auth"
	"github.com/boxingoctopus/snackmates/api/internal/cache"
	"github.com/boxingoctopus/snackmates/api/internal/config"
	"github.com/boxingoctopus/snackmates/api/internal/middleware"
	"github.com/boxingoctopus/snackmates/api/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthHandler struct {
	cfg      config.Config
	pool     *pgxpool.Pool
	cache    *cache.Client
	storage  *storage.Client
	discord  *auth.DiscordService
	webauthn *auth.WebAuthnService
}

func NewAuthHandler(cfg config.Config, pool *pgxpool.Pool, c *cache.Client, s *storage.Client, discord *auth.DiscordService, wa *auth.WebAuthnService) *AuthHandler {
	return &AuthHandler{cfg: cfg, pool: pool, cache: c, storage: s, discord: discord, webauthn: wa}
}

func (h *AuthHandler) RegisterRoutes(app fiber.Router) {
	auth := app.Group("/auth")
	auth.Post("/register", h.Register)
	auth.Post("/login", h.Login)
	auth.Post("/logout", h.Logout)
	auth.Post("/verify-email", h.VerifyEmail)
	auth.Post("/forgot-password", h.ForgotPassword)
	auth.Post("/reset-password", h.ResetPassword)
	account := auth.Group("/account")
	account.Post("/deactivate/request", middleware.RequireAuth(h.pool), h.RequestAccountDeactivate)
	account.Post("/delete/request", middleware.RequireAuth(h.pool), h.RequestAccountDelete)
	account.Post("/reactivate/request", h.RequestAccountReactivate)
	account.Post("/confirm", h.ConfirmAccountAction)
	auth.Get("/discord", h.DiscordStart)
	auth.Get("/discord/connect", middleware.RequireAuth(h.pool), h.DiscordConnect)
	auth.Get("/discord/callback", h.DiscordCallback)
	auth.Get("/me", middleware.RequireAuth(h.pool), h.Me)

	mfa := auth.Group("/mfa", middleware.RequireAuth(h.pool))
	mfa.Post("/totp/setup", h.TOTPSetup)
	mfa.Post("/totp/enable", h.TOTPEnable)
	mfa.Post("/totp/verify", h.TOTPVerify)
	mfa.Post("/totp/disable", h.TOTPDisable)
	mfa.Post("/webauthn/register/begin", h.WebAuthnRegisterBegin)
	mfa.Post("/webauthn/register/finish", h.WebAuthnRegisterFinishRaw)
	mfa.Post("/webauthn/login/begin", h.WebAuthnLoginBegin)
	mfa.Post("/webauthn/login/finish", h.WebAuthnLoginFinishRaw)
}

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
	Country     string `json:"country"`
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if req.Email == "" || req.Password == "" || req.DisplayName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email, password, and display_name required"})
	}
	userID, err := auth.Register(c.Context(), h.pool, h.cfg, req.Email, req.Password, req.DisplayName, req.Country, req.Username)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"user_id": userID, "message": "check your email to verify your account"})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TOTPCode string `json:"totp_code"`
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	userID, token, totpRequired, err := auth.Login(c.Context(), h.pool, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrAccountDeactivated) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "This account has been deactivated. Check your email for a reactivation link.",
				"code":  "account_deactivated",
			})
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}
	if totpRequired {
		if req.TOTPCode == "" {
			return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"mfa_required": true, "methods": []string{"totp", "webauthn"}})
		}
		if err := auth.ValidateTOTP(c.Context(), h.pool, userID, req.TOTPCode); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid totp code"})
		}
	}
	setSessionCookie(c, token)
	return c.JSON(fiber.Map{"token": token})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	token := c.Cookies("session")
	if ah := c.Get("Authorization"); len(ah) > 7 {
		token = ah[7:]
	}
	if token != "" {
		_ = auth.Logout(c.Context(), h.pool, token)
	}
	c.ClearCookie("session")
	return c.JSON(fiber.Map{"ok": true})
}

func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	var body struct {
		Token string `json:"token"`
	}
	if err := c.BodyParser(&body); err != nil || body.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "token required"})
	}
	if err := auth.VerifyEmail(c.Context(), h.pool, body.Token); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var body struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	_ = auth.RequestPasswordReset(c.Context(), h.pool, h.cfg, body.Email)
	return c.JSON(fiber.Map{"message": "if that email exists, a reset link was sent"})
}

func (h *AuthHandler) RequestAccountDeactivate(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if err := auth.RequestAccountDeactivate(c.Context(), h.pool, h.cfg, userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Check your email to confirm account deactivation."})
}

func (h *AuthHandler) RequestAccountDelete(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if err := auth.RequestAccountDelete(c.Context(), h.pool, h.cfg, userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Check your email to confirm account deletion."})
}

func (h *AuthHandler) RequestAccountReactivate(c *fiber.Ctx) error {
	var body struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&body); err != nil || body.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email required"})
	}
	_ = auth.RequestAccountReactivate(c.Context(), h.pool, h.cfg, body.Email)
	return c.JSON(fiber.Map{"message": "If that account is deactivated, a reactivation link was sent."})
}

func (h *AuthHandler) ConfirmAccountAction(c *fiber.Ctx) error {
	var body struct {
		Token string `json:"token"`
	}
	if err := c.BodyParser(&body); err != nil || body.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "token required"})
	}
	action, err := auth.ConfirmAccountAction(c.Context(), h.pool, body.Token)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"ok": true, "action": action})
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var body struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := auth.ResetPassword(c.Context(), h.pool, body.Token, body.NewPassword); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	rec, err := auth.GetUserByID(c.Context(), h.pool, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	return c.JSON(userProfileResponse(c.Context(), h.storage, rec))
}

func (h *AuthHandler) DiscordStart(c *fiber.Ctx) error {
	if !h.discord.Enabled() {
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"error": "discord oauth not configured"})
	}
	state, err := auth.NewToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to start oauth"})
	}
	if err := auth.StoreOAuthState(h.cache, state, auth.OAuthPurposeLogin()); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to start oauth"})
	}
	return c.Redirect(h.discord.AuthURL(state))
}

func (h *AuthHandler) DiscordConnect(c *fiber.Ctx) error {
	if !h.discord.Enabled() {
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"error": "discord oauth not configured"})
	}
	userID := middleware.GetUserID(c)
	state, err := auth.NewToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to start oauth"})
	}
	if err := auth.StoreOAuthState(h.cache, state, auth.ConnectOAuthPurpose(userID)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to start oauth"})
	}
	return c.JSON(fiber.Map{"url": h.discord.AuthURL(state)})
}

func (h *AuthHandler) DiscordCallback(c *fiber.Ctx) error {
	if !h.discord.Enabled() {
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"error": "discord oauth not configured"})
	}
	state := c.Query("state")
	purpose, err := auth.ConsumeOAuthState(h.cache, state)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid oauth state"})
	}
	code := c.Query("code")

	if auth.IsConnectOAuthPurpose(purpose) {
		userID, err := auth.UserIDFromConnectPurpose(purpose)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid oauth state"})
		}
		if err := h.discord.LinkAccount(c.Context(), h.pool, code, userID); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		rec, err := auth.GetUserByID(c.Context(), h.pool, userID)
		if err != nil {
			return c.Redirect(h.cfg.WebOrigin + "/settings?discord=connected")
		}
		return c.Redirect(h.cfg.WebOrigin + "/users/" + url.PathEscape(rec.Username) + "?discord=connected")
	}

	_, token, err := h.discord.HandleLoginCallback(c.Context(), h.pool, code)
	if err != nil {
		if errors.Is(err, auth.ErrAccountDeactivated) {
			return c.Redirect(h.cfg.WebOrigin + "/login?error=" + url.QueryEscape("This account has been deactivated. Check your email for a reactivation link."))
		}
		return c.Redirect(h.cfg.WebOrigin + "/login?error=" + url.QueryEscape(err.Error()))
	}
	setSessionCookie(c, token)
	return c.Redirect(h.cfg.WebOrigin + "/dashboard?token=" + url.QueryEscape(token))
}

func (h *AuthHandler) TOTPSetup(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	secret, url, err := auth.SetupTOTP(c.Context(), h.pool, userID, "SnackMates")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"secret": secret, "otpauth_url": url})
}

func (h *AuthHandler) TOTPEnable(c *fiber.Ctx) error {
	var body struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := auth.EnableTOTP(c.Context(), h.pool, middleware.GetUserID(c), body.Code); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *AuthHandler) TOTPVerify(c *fiber.Ctx) error {
	var body struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := auth.ValidateTOTP(c.Context(), h.pool, middleware.GetUserID(c), body.Code); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *AuthHandler) TOTPDisable(c *fiber.Ctx) error {
	var body struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := auth.DisableTOTP(c.Context(), h.pool, middleware.GetUserID(c), body.Code); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *AuthHandler) WebAuthnRegisterBegin(c *fiber.Ctx) error {
	resp, err := h.webauthn.BeginRegistration(c.Context(), h.pool, middleware.GetUserID(c))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(resp)
}

func (h *AuthHandler) WebAuthnLoginBegin(c *fiber.Ctx) error {
	resp, err := h.webauthn.BeginLogin(c.Context(), h.pool, middleware.GetUserID(c))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(resp)
}

func setSessionCookie(c *fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     "session",
		Value:    token,
		HTTPOnly: true,
		SameSite: "Lax",
		Path:     "/",
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
	})
}
