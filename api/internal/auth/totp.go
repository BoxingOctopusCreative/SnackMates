package auth

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pquerna/otp/totp"
)

func SetupTOTP(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, issuer string) (secret string, qrURL string, err error) {
	rec, err := GetUserByID(ctx, pool, userID)
	if err != nil {
		return "", "", err
	}
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: rec.Email,
	})
	if err != nil {
		return "", "", err
	}
	_, err = pool.Exec(ctx, `UPDATE users SET totp_secret = $2, totp_enabled = FALSE WHERE id = $1`, userID, key.Secret())
	if err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

func EnableTOTP(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, code string) error {
	var secret *string
	err := pool.QueryRow(ctx, `SELECT totp_secret FROM users WHERE id = $1`, userID).Scan(&secret)
	if err != nil || secret == nil || *secret == "" {
		return fmt.Errorf("totp not configured")
	}
	if !totp.Validate(code, *secret) {
		return fmt.Errorf("invalid totp code")
	}
	_, err = pool.Exec(ctx, `UPDATE users SET totp_enabled = TRUE WHERE id = $1`, userID)
	return err
}

func ValidateTOTP(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, code string) error {
	var secret *string
	var enabled bool
	err := pool.QueryRow(ctx, `SELECT totp_secret, totp_enabled FROM users WHERE id = $1`, userID).Scan(&secret, &enabled)
	if err != nil {
		return err
	}
	if !enabled || secret == nil {
		return fmt.Errorf("totp not enabled")
	}
	if !totp.Validate(code, *secret) {
		return fmt.Errorf("invalid totp code")
	}
	return nil
}

func DisableTOTP(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, code string) error {
	if err := ValidateTOTP(ctx, pool, userID, code); err != nil {
		return err
	}
	_, err := pool.Exec(ctx, `UPDATE users SET totp_enabled = FALSE, totp_secret = NULL WHERE id = $1`, userID)
	return err
}

func TOTPQRDataURL(url string) string {
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte(url))
}
