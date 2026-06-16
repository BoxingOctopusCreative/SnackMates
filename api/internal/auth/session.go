package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const rememberedSessionThreshold = 48 * time.Hour

// SessionTTL returns how long a new session should remain valid.
func SessionTTL(remember bool) time.Duration {
	if remember {
		return 30 * 24 * time.Hour
	}
	return 24 * time.Hour
}

// CookieMaxAge returns the Max-Age for a session cookie. Short-lived sessions use
// browser session cookies (MaxAge 0).
func CookieMaxAge(expiresAt time.Time) int {
	if time.Until(expiresAt) > rememberedSessionThreshold {
		return int(time.Until(expiresAt).Seconds())
	}
	return 0
}

// SessionRemembered reports whether a session should persist across browser restarts.
func SessionRemembered(expiresAt time.Time) bool {
	return time.Until(expiresAt) > rememberedSessionThreshold
}

func CreateSession(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, remember bool) (string, time.Time, error) {
	token, err := NewToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(SessionTTL(remember))
	_, err = pool.Exec(ctx, `
		INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, HashToken(token), expiresAt)
	return token, expiresAt, err
}

func SessionDetails(ctx context.Context, pool *pgxpool.Pool, token string) (uuid.UUID, time.Time, error) {
	var userID uuid.UUID
	var expiresAt time.Time
	err := pool.QueryRow(ctx, `
		SELECT s.user_id, s.expires_at FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.expires_at > NOW() AND u.deactivated_at IS NULL
	`, HashToken(token)).Scan(&userID, &expiresAt)
	return userID, expiresAt, err
}
