package matching

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/boxingoctopus/snackmates/api/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type eligibleUser struct {
	id      uuid.UUID
	country string
}

// PairUsers selects eligible users and pairs them at random across different countries.
// Users must have a country set, at least one public wishlist item, and no active match.
func PairUsers(ctx context.Context, pool *pgxpool.Pool) ([]models.SnackMatch, error) {
	rows, err := pool.Query(ctx, `
		SELECT u.id, UPPER(TRIM(u.country))
		FROM users u
		WHERE u.email_verified = TRUE
		  AND u.deactivated_at IS NULL
		  AND NULLIF(TRIM(u.country), '') IS NOT NULL
		  AND EXISTS (
		    SELECT 1 FROM wishlists w
		    JOIN wishlist_items wi ON wi.wishlist_id = w.id
		    WHERE w.user_id = u.id AND w.is_public = TRUE
		  )
		  AND NOT EXISTS (
		    SELECT 1 FROM snack_matches sm
		    WHERE (sm.user_a_id = u.id OR sm.user_b_id = u.id)
		      AND sm.status IN ('pending', 'active')
		  )
	`)
	if err != nil {
		return nil, fmt.Errorf("query eligible users: %w", err)
	}
	defer rows.Close()

	var users []eligibleUser
	for rows.Next() {
		var u eligibleUser
		if err := rows.Scan(&u.id, &u.country); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	pairs := pairCrossCountry(users, rng)

	matches := make([]models.SnackMatch, 0)
	for _, pair := range pairs {
		var match models.SnackMatch
		err := pool.QueryRow(ctx, `
			INSERT INTO snack_matches (user_a_id, user_b_id, status, expires_at)
			VALUES ($1, $2, 'active', NOW() + INTERVAL '30 days')
			ON CONFLICT (user_a_id, user_b_id) DO NOTHING
			RETURNING id, user_a_id, user_b_id, status, matched_at
		`, pair[0], pair[1]).Scan(&match.ID, &match.UserAID, &match.UserBID, &match.Status, &match.MatchedAt)
		if err != nil {
			continue
		}
		matches = append(matches, match)
	}
	return matches, nil
}

func pairCrossCountry(users []eligibleUser, rng *rand.Rand) [][2]uuid.UUID {
	if len(users) < 2 {
		return nil
	}

	remaining := append([]eligibleUser(nil), users...)
	rng.Shuffle(len(remaining), func(i, j int) {
		remaining[i], remaining[j] = remaining[j], remaining[i]
	})

	var pairs [][2]uuid.UUID
	for len(remaining) > 0 {
		current := remaining[0]
		remaining = remaining[1:]

		partnerIdx := -1
		for i, candidate := range remaining {
			if !sameCountry(current.country, candidate.country) {
				partnerIdx = i
				break
			}
		}
		if partnerIdx < 0 {
			continue
		}

		partner := remaining[partnerIdx]
		remaining = append(remaining[:partnerIdx], remaining[partnerIdx+1:]...)

		a, b := orderUsers(current.id, partner.id)
		pairs = append(pairs, [2]uuid.UUID{a, b})
	}
	return pairs
}

func sameCountry(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

func orderUsers(a, b uuid.UUID) (uuid.UUID, uuid.UUID) {
	if a.String() > b.String() {
		return b, a
	}
	return a, b
}

func GetMatchesForUser(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) ([]models.SnackMatch, error) {
	rows, err := pool.Query(ctx, `
		SELECT sm.id, sm.user_a_id, sm.user_b_id, sm.status, sm.matched_at,
		       u.id, u.username, u.email, u.email_verified, u.display_name, u.bio, u.country,
		       u.avatar_key, u.avatar_url, u.discord_id, u.totp_enabled, u.created_at,
		       EXISTS(SELECT 1 FROM webauthn_credentials wc WHERE wc.user_id = u.id)
		FROM snack_matches sm
		JOIN users u ON u.id = CASE WHEN sm.user_a_id = $1 THEN sm.user_b_id ELSE sm.user_a_id END
		WHERE sm.user_a_id = $1 OR sm.user_b_id = $1
		ORDER BY sm.matched_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches := make([]models.SnackMatch, 0)
	for rows.Next() {
		var m models.SnackMatch
		var mate models.User
		if err := rows.Scan(
			&m.ID, &m.UserAID, &m.UserBID, &m.Status, &m.MatchedAt,
			&mate.ID, &mate.Username, &mate.Email, &mate.EmailVerified, &mate.DisplayName, &mate.Bio, &mate.Country,
			&mate.AvatarKey, &mate.AvatarURL, &mate.DiscordID, &mate.TOTPEnabled, &mate.CreatedAt, &mate.HasWebAuthn,
		); err != nil {
			return nil, err
		}
		m.Mate = &mate
		matches = append(matches, m)
	}
	return matches, rows.Err()
}
