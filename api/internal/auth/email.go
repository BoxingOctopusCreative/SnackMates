package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/boxingoctopus/snackmates/api/internal/config"
	mail "github.com/boxingoctopus/snackmates/api/internal/email"
	"github.com/boxingoctopus/snackmates/api/internal/slug"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func NewToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func Register(ctx context.Context, pool *pgxpool.Pool, cfg config.Config, email, password, displayName, country, username string) (*uuid.UUID, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(username) == "" {
		username = slug.UsernameFromName(displayName)
	} else {
		username = strings.ToLower(strings.TrimSpace(username))
		if !slug.IsValidUsername(username) {
			return nil, fmt.Errorf("username must be 3-32 lowercase letters or numbers")
		}
	}
	username, err = allocateUsername(ctx, pool, username)
	if err != nil {
		return nil, err
	}
	var userID uuid.UUID
	err = pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, display_name, country, username)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, email, hash, displayName, country, username).Scan(&userID)
	if err != nil {
		return nil, err
	}

	token, err := NewToken()
	if err != nil {
		return nil, err
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO email_verification_tokens (user_id, token, expires_at)
		VALUES ($1, $2, NOW() + INTERVAL '24 hours')
	`, userID, HashToken(token))
	if err != nil {
		return nil, err
	}

	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", cfg.WebOrigin, token)
	body := fmt.Sprintf("<p>Welcome to SnackMates! <a href=\"%s\">Verify your email</a> to start matching snack mates.</p>", verifyURL)
	_ = mail.Send(cfg, email, "Verify your SnackMates account", body)
	return &userID, nil
}

func VerifyEmail(ctx context.Context, pool *pgxpool.Pool, token string) error {
	tokenHash := HashToken(token)
	tag, err := pool.Exec(ctx, `
		UPDATE users u SET email_verified = TRUE
		FROM email_verification_tokens t
		WHERE t.token = $1 AND t.expires_at > NOW() AND u.id = t.user_id
	`, tokenHash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("invalid or expired token")
	}
	_, _ = pool.Exec(ctx, `DELETE FROM email_verification_tokens WHERE token = $1`, tokenHash)
	return nil
}

func Login(ctx context.Context, pool *pgxpool.Pool, email, password string) (uuid.UUID, string, bool, error) {
	var userID uuid.UUID
	var hash string
	var totpEnabled bool
	var deactivatedAt *time.Time
	err := pool.QueryRow(ctx, `
		SELECT id, password_hash, totp_enabled, deactivated_at FROM users WHERE email = $1
	`, email).Scan(&userID, &hash, &totpEnabled, &deactivatedAt)
	if err != nil {
		return uuid.Nil, "", false, fmt.Errorf("invalid credentials")
	}
	if hash == "" || !CheckPassword(hash, password) {
		return uuid.Nil, "", false, fmt.Errorf("invalid credentials")
	}
	if deactivatedAt != nil {
		return uuid.Nil, "", false, ErrAccountDeactivated
	}
	sessionToken, err := CreateSession(ctx, pool, userID)
	if err != nil {
		return uuid.Nil, "", false, err
	}
	return userID, sessionToken, totpEnabled, nil
}

func CreateSession(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) (string, error) {
	token, err := NewToken()
	if err != nil {
		return "", err
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, NOW() + INTERVAL '7 days')
	`, userID, HashToken(token))
	return token, err
}

func ValidateSession(ctx context.Context, pool *pgxpool.Pool, token string) (uuid.UUID, error) {
	var userID uuid.UUID
	err := pool.QueryRow(ctx, `
		SELECT s.user_id FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.expires_at > NOW() AND u.deactivated_at IS NULL
	`, HashToken(token)).Scan(&userID)
	return userID, err
}

func Logout(ctx context.Context, pool *pgxpool.Pool, token string) error {
	_, err := pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, HashToken(token))
	return err
}

func RequestPasswordReset(ctx context.Context, pool *pgxpool.Pool, cfg config.Config, email string) error {
	var userID uuid.UUID
	err := pool.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&userID)
	if err != nil {
		return nil
	}
	token, err := NewToken()
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO password_reset_tokens (user_id, token, expires_at)
		VALUES ($1, $2, NOW() + INTERVAL '1 hour')
	`, userID, HashToken(token))
	if err != nil {
		return err
	}
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", cfg.WebOrigin, token)
	body := fmt.Sprintf("<p>Reset your SnackMates password: <a href=\"%s\">Click here</a></p>", resetURL)
	return mail.Send(cfg, email, "Reset your SnackMates password", body)
}

func ResetPassword(ctx context.Context, pool *pgxpool.Pool, token, newPassword string) error {
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	tokenHash := HashToken(token)
	tag, err := pool.Exec(ctx, `
		UPDATE users u SET password_hash = $2
		FROM password_reset_tokens t
		WHERE t.token = $1 AND t.expires_at > NOW() AND u.id = t.user_id
	`, tokenHash, hash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("invalid or expired token")
	}
	_, _ = pool.Exec(ctx, `DELETE FROM password_reset_tokens WHERE token = $1`, tokenHash)
	return nil
}

func GetUserByID(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) (*UserRecord, error) {
	var u UserRecord
	err := pool.QueryRow(ctx, `
		SELECT u.id, u.email, u.email_verified, u.username, u.display_name, u.bio, u.country,
		       u.avatar_key, u.avatar_url, u.banner_key, u.banner_url,
		       u.discord_id, u.totp_enabled, u.created_at, u.deactivated_at,
		       EXISTS(SELECT 1 FROM webauthn_credentials wc WHERE wc.user_id = u.id)
		FROM users u WHERE u.id = $1
	`, userID).Scan(
		&u.ID, &u.Email, &u.EmailVerified, &u.Username, &u.DisplayName, &u.Bio, &u.Country,
		&u.AvatarKey, &u.AvatarURL, &u.BannerKey, &u.BannerURL,
		&u.DiscordID, &u.TOTPEnabled, &u.CreatedAt, &u.DeactivatedAt, &u.HasWebAuthn,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func GetUserByUsername(ctx context.Context, pool *pgxpool.Pool, username string) (*UserRecord, error) {
	var u UserRecord
	err := pool.QueryRow(ctx, `
		SELECT u.id, u.email, u.email_verified, u.username, u.display_name, u.bio, u.country,
		       u.avatar_key, u.avatar_url, u.banner_key, u.banner_url,
		       u.discord_id, u.totp_enabled, u.created_at, u.deactivated_at,
		       EXISTS(SELECT 1 FROM webauthn_credentials wc WHERE wc.user_id = u.id)
		FROM users u WHERE lower(u.username) = lower($1)
	`, username).Scan(
		&u.ID, &u.Email, &u.EmailVerified, &u.Username, &u.DisplayName, &u.Bio, &u.Country,
		&u.AvatarKey, &u.AvatarURL, &u.BannerKey, &u.BannerURL,
		&u.DiscordID, &u.TOTPEnabled, &u.CreatedAt, &u.DeactivatedAt, &u.HasWebAuthn,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func allocateUsername(ctx context.Context, pool *pgxpool.Pool, base string) (string, error) {
	base = slug.UsernameFromName(base)
	return slug.Unique(base, func(candidate string) bool {
		var exists bool
		err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE lower(username) = lower($1))`, candidate).Scan(&exists)
		return err != nil || exists
	}), nil
}

type UserRecord struct {
	ID            uuid.UUID
	Email         string
	EmailVerified bool
	Username      string
	DisplayName   string
	Bio           string
	Country       string
	AvatarKey     *string
	AvatarURL     *string
	BannerKey     *string
	BannerURL     *string
	DiscordID     *string
	TOTPEnabled   bool
	HasWebAuthn   bool
	CreatedAt     time.Time
	DeactivatedAt *time.Time
}

func (u UserRecord) ToModel() interface{} {
	return map[string]interface{}{
		"id":              u.ID,
		"username":        u.Username,
		"email":           u.Email,
		"email_verified":  u.EmailVerified,
		"display_name":    u.DisplayName,
		"bio":             u.Bio,
		"country":         u.Country,
		"avatar_key":      u.AvatarKey,
		"avatar_url":      u.AvatarURL,
		"banner_key":      u.BannerKey,
		"banner_url":      u.BannerURL,
		"discord_id":      u.DiscordID,
		"discord_linked":  u.DiscordID != nil && *u.DiscordID != "",
		"totp_enabled":    u.TOTPEnabled,
		"has_webauthn":    u.HasWebAuthn,
		"created_at":      u.CreatedAt,
	}
}
