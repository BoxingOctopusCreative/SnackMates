package search

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserHit struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Bio         string    `json:"bio"`
	Country     string    `json:"country"`
	AvatarKey   *string   `json:"-"`
	AvatarURL   *string   `json:"-"`
}

func SearchUsers(ctx context.Context, pool *pgxpool.Pool, query string, limit int) ([]UserHit, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []UserHit{}, nil
	}
	if limit <= 0 {
		limit = 8
	}

	pattern := "%" + query + "%"
	rows, err := pool.Query(ctx, `
		SELECT u.id, u.username, u.display_name, COALESCE(u.bio, ''), COALESCE(u.country, ''),
		       u.avatar_key, u.avatar_url
		FROM users u
		WHERE u.email_verified = TRUE
		  AND u.deactivated_at IS NULL
		  AND (
		    u.username ILIKE $1
		    OR u.display_name ILIKE $1
		    OR COALESCE(u.bio, '') ILIKE $1
		  )
		ORDER BY
		  CASE
		    WHEN lower(u.username) = lower($2) THEN 0
		    WHEN lower(u.display_name) = lower($2) THEN 1
		    WHEN u.username ILIKE $3 THEN 2
		    WHEN u.display_name ILIKE $3 THEN 3
		    ELSE 4
		  END,
		  u.display_name ASC
		LIMIT $4
	`, pattern, query, query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("search users: %w", err)
	}
	defer rows.Close()

	users := make([]UserHit, 0)
	for rows.Next() {
		var u UserHit
		if err := rows.Scan(
			&u.ID, &u.Username, &u.DisplayName, &u.Bio, &u.Country,
			&u.AvatarKey, &u.AvatarURL,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
