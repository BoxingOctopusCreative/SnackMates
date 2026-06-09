package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/boxingoctopus/snackmates/api/internal/config"
	mail "github.com/boxingoctopus/snackmates/api/internal/email"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAccountDeactivated = errors.New("account_deactivated")

type AccountAction string

const (
	AccountActionDeactivate AccountAction = "deactivate"
	AccountActionDelete     AccountAction = "delete"
	AccountActionReactivate AccountAction = "reactivate"
)

func RequestAccountDeactivate(ctx context.Context, pool *pgxpool.Pool, cfg config.Config, userID uuid.UUID) error {
	return requestAccountAction(ctx, pool, cfg, userID, AccountActionDeactivate)
}

func RequestAccountDelete(ctx context.Context, pool *pgxpool.Pool, cfg config.Config, userID uuid.UUID) error {
	return requestAccountAction(ctx, pool, cfg, userID, AccountActionDelete)
}

func RequestAccountReactivate(ctx context.Context, pool *pgxpool.Pool, cfg config.Config, email string) error {
	var userID uuid.UUID
	err := pool.QueryRow(ctx, `
		SELECT id FROM users
		WHERE email = $1 AND deactivated_at IS NOT NULL
	`, email).Scan(&userID)
	if err != nil {
		return nil
	}
	return requestAccountAction(ctx, pool, cfg, userID, AccountActionReactivate)
}

func requestAccountAction(ctx context.Context, pool *pgxpool.Pool, cfg config.Config, userID uuid.UUID, action AccountAction) error {
	var email string
	var deactivatedAt interface{}
	err := pool.QueryRow(ctx, `
		SELECT email, deactivated_at FROM users WHERE id = $1
	`, userID).Scan(&email, &deactivatedAt)
	if err != nil {
		return err
	}

	switch action {
	case AccountActionDeactivate, AccountActionDelete:
		if deactivatedAt != nil {
			return fmt.Errorf("account is already deactivated")
		}
	case AccountActionReactivate:
		if deactivatedAt == nil {
			return nil
		}
	}

	token, err := NewToken()
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `DELETE FROM account_action_tokens WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO account_action_tokens (user_id, action, token, expires_at)
		VALUES ($1, $2, $3, NOW() + INTERVAL '24 hours')
	`, userID, string(action), HashToken(token))
	if err != nil {
		return err
	}

	confirmURL := fmt.Sprintf("%s/confirm-account?token=%s", cfg.WebOrigin, token)
	subject, body := accountActionEmail(action, confirmURL)
	return mail.Send(cfg, email, subject, body)
}

func ConfirmAccountAction(ctx context.Context, pool *pgxpool.Pool, token string) (AccountAction, error) {
	tokenHash := HashToken(token)
	var userID uuid.UUID
	var action AccountAction
	err := pool.QueryRow(ctx, `
		SELECT user_id, action::text
		FROM account_action_tokens
		WHERE token = $1 AND expires_at > NOW()
	`, tokenHash).Scan(&userID, &action)
	if err != nil {
		return "", fmt.Errorf("invalid or expired token")
	}

	switch action {
	case AccountActionDeactivate:
		if err := deactivateUser(ctx, pool, userID); err != nil {
			return "", err
		}
	case AccountActionDelete:
		if err := deleteUser(ctx, pool, userID); err != nil {
			return "", err
		}
	case AccountActionReactivate:
		if err := reactivateUser(ctx, pool, userID); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unsupported account action")
	}

	_, _ = pool.Exec(ctx, `DELETE FROM account_action_tokens WHERE user_id = $1`, userID)
	return action, nil
}

func deactivateUser(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) error {
	tag, err := pool.Exec(ctx, `
		UPDATE users
		SET deactivated_at = NOW()
		WHERE id = $1 AND deactivated_at IS NULL
	`, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("account is already deactivated")
	}

	_, _ = pool.Exec(ctx, `
		UPDATE snack_matches
		SET status = 'cancelled'
		WHERE status IN ('pending', 'active')
		  AND (user_a_id = $1 OR user_b_id = $1)
	`, userID)

	return revokeAllSessions(ctx, pool, userID)
}

func reactivateUser(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) error {
	tag, err := pool.Exec(ctx, `
		UPDATE users
		SET deactivated_at = NULL
		WHERE id = $1 AND deactivated_at IS NOT NULL
	`, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("account is not deactivated")
	}
	return nil
}

func deleteUser(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) error {
	tag, err := pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func revokeAllSessions(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) error {
	_, err := pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

func accountActionEmail(action AccountAction, confirmURL string) (string, string) {
	switch action {
	case AccountActionDeactivate:
		return "Confirm SnackMates Account Deactivation",
			fmt.Sprintf("<p>We received a request to deactivate your SnackMates account. <a href=\"%s\">Confirm deactivation</a>. This link expires in 24 hours.</p><p>If you did not request this, you can ignore this email.</p>", confirmURL)
	case AccountActionDelete:
		return "Confirm SnackMates Account Deletion",
			fmt.Sprintf("<p>We received a request to permanently delete your SnackMates account. <a href=\"%s\">Confirm deletion</a>. This cannot be undone. This link expires in 24 hours.</p><p>If you did not request this, you can ignore this email.</p>", confirmURL)
	default:
		return "Reactivate Your SnackMates Account",
			fmt.Sprintf("<p>We received a request to reactivate your SnackMates account. <a href=\"%s\">Confirm reactivation</a>. This link expires in 24 hours.</p>", confirmURL)
	}
}
