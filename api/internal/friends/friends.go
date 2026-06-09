package friends

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/boxingoctopus/snackmates/api/internal/auth"
	"github.com/boxingoctopus/snackmates/api/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrSelfFriend       = errors.New("cannot friend yourself")
	ErrUserNotFound     = errors.New("user not found")
	ErrAlreadyFriends   = errors.New("already friends")
	ErrRequestPending    = errors.New("snack mate request already sent")
	ErrIncomingRequest   = errors.New("accept their snack mate request first")
	ErrFriendshipNotFound = errors.New("friendship not found")
	ErrNotAuthorized    = errors.New("not authorized")
)

type Record struct {
	ID          uuid.UUID
	RequesterID uuid.UUID
	AddresseeID uuid.UUID
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func LookupBetween(ctx context.Context, pool *pgxpool.Pool, userA, userB uuid.UUID) (*Record, error) {
	var rec Record
	err := pool.QueryRow(ctx, `
		SELECT id, requester_id, addressee_id, status::text, created_at, updated_at
		FROM friendships
		WHERE (requester_id = $1 AND addressee_id = $2)
		   OR (requester_id = $2 AND addressee_id = $1)
	`, userA, userB).Scan(
		&rec.ID, &rec.RequesterID, &rec.AddresseeID, &rec.Status, &rec.CreatedAt, &rec.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func ViewFor(viewerID, profileUserID uuid.UUID, rec *Record) *models.FriendshipView {
	if viewerID == uuid.Nil || viewerID == profileUserID || rec == nil {
		return nil
	}

	view := &models.FriendshipView{ID: &rec.ID}
	switch rec.Status {
	case "accepted":
		view.Status = "friends"
	case "declined":
		view.Status = "declined"
	case "pending":
		if rec.RequesterID == viewerID {
			view.Status = "pending_outgoing"
		} else {
			view.Status = "pending_incoming"
		}
	default:
		return nil
	}
	return view
}

func Request(ctx context.Context, pool *pgxpool.Pool, requesterID uuid.UUID, username string) (*Record, error) {
	target, err := auth.GetUserByUsername(ctx, pool, username)
	if err != nil || !target.EmailVerified || target.DeactivatedAt != nil {
		return nil, ErrUserNotFound
	}
	if target.ID == requesterID {
		return nil, ErrSelfFriend
	}

	existing, err := LookupBetween(ctx, pool, requesterID, target.ID)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		switch existing.Status {
		case "accepted":
			return nil, ErrAlreadyFriends
		case "pending":
			if existing.RequesterID == requesterID {
				return nil, ErrRequestPending
			}
			return nil, ErrIncomingRequest
		case "declined":
			return updateRequest(ctx, pool, existing.ID, requesterID, target.ID)
		}
	}

	return insert(ctx, pool, requesterID, target.ID)
}

func Accept(ctx context.Context, pool *pgxpool.Pool, friendshipID, userID uuid.UUID) (*Record, error) {
	rec, err := getByID(ctx, pool, friendshipID)
	if err != nil {
		return nil, err
	}
	if rec.Status != "pending" || rec.AddresseeID != userID {
		return nil, ErrNotAuthorized
	}
	return setStatus(ctx, pool, friendshipID, "accepted")
}

func Decline(ctx context.Context, pool *pgxpool.Pool, friendshipID, userID uuid.UUID) (*Record, error) {
	rec, err := getByID(ctx, pool, friendshipID)
	if err != nil {
		return nil, err
	}
	if rec.Status != "pending" || rec.AddresseeID != userID {
		return nil, ErrNotAuthorized
	}
	updated, err := setStatus(ctx, pool, friendshipID, "declined")
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func Remove(ctx context.Context, pool *pgxpool.Pool, friendshipID, userID uuid.UUID) (*Record, error) {
	rec, err := getByID(ctx, pool, friendshipID)
	if err != nil {
		return nil, err
	}
	switch rec.Status {
	case "accepted":
		if rec.RequesterID != userID && rec.AddresseeID != userID {
			return nil, ErrNotAuthorized
		}
	case "pending":
		if rec.RequesterID != userID {
			return nil, ErrNotAuthorized
		}
	default:
		return nil, ErrNotAuthorized
	}

	tag, err := pool.Exec(ctx, `DELETE FROM friendships WHERE id = $1`, friendshipID)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrFriendshipNotFound
	}
	return rec, nil
}

func ListAccepted(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) ([]models.Friendship, error) {
	rows, err := pool.Query(ctx, `
		SELECT f.id, f.requester_id, f.addressee_id, f.status::text, f.created_at, f.updated_at,
		       u.id, u.username, u.display_name, COALESCE(u.bio, ''), COALESCE(u.country, ''),
		       u.avatar_key, u.avatar_url
		FROM friendships f
		JOIN users u ON u.id = CASE
			WHEN f.requester_id = $1 THEN f.addressee_id
			ELSE f.requester_id
		END
		WHERE f.status = 'accepted'
		  AND (f.requester_id = $1 OR f.addressee_id = $1)
		  AND u.email_verified = TRUE
		  AND u.deactivated_at IS NULL
		ORDER BY f.updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFriendships(rows)
}

func ListAcceptedUnseenForRequester(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) ([]models.Friendship, error) {
	rows, err := pool.Query(ctx, `
		SELECT f.id, f.requester_id, f.addressee_id, f.status::text, f.created_at, f.updated_at,
		       u.id, u.username, u.display_name, COALESCE(u.bio, ''), COALESCE(u.country, ''),
		       u.avatar_key, u.avatar_url
		FROM friendships f
		JOIN users u ON u.id = f.addressee_id
		WHERE f.requester_id = $1
		  AND f.status = 'accepted'
		  AND f.requester_notified_at IS NULL
		  AND u.email_verified = TRUE
		  AND u.deactivated_at IS NULL
		ORDER BY f.updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFriendships(rows)
}

func AcknowledgeAcceptance(ctx context.Context, pool *pgxpool.Pool, friendshipID, requesterID uuid.UUID) error {
	tag, err := pool.Exec(ctx, `
		UPDATE friendships
		SET requester_notified_at = NOW()
		WHERE id = $1
		  AND requester_id = $2
		  AND status = 'accepted'
		  AND requester_notified_at IS NULL
	`, friendshipID, requesterID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrFriendshipNotFound
	}
	return nil
}

func ListIncoming(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) ([]models.Friendship, error) {
	rows, err := pool.Query(ctx, `
		SELECT f.id, f.requester_id, f.addressee_id, f.status::text, f.created_at, f.updated_at,
		       u.id, u.username, u.display_name, COALESCE(u.bio, ''), COALESCE(u.country, ''),
		       u.avatar_key, u.avatar_url
		FROM friendships f
		JOIN users u ON u.id = f.requester_id
		WHERE f.addressee_id = $1
		  AND f.status = 'pending'
		  AND u.email_verified = TRUE
		  AND u.deactivated_at IS NULL
		ORDER BY f.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFriendships(rows)
}

func scanFriendships(rows pgx.Rows) ([]models.Friendship, error) {
	list := make([]models.Friendship, 0)
	for rows.Next() {
		var f models.Friendship
		var user models.PublicUser
		var avatarKey, avatarURL *string
		if err := rows.Scan(
			&f.ID, &f.RequesterID, &f.AddresseeID, &f.Status, &f.CreatedAt, &f.UpdatedAt,
			&user.ID, &user.Username, &user.DisplayName, &user.Bio, &user.Country,
			&avatarKey, &avatarURL,
		); err != nil {
			return nil, err
		}
		if avatarURL != nil && *avatarURL != "" {
			user.AvatarURL = avatarURL
		}
		f.User = &user
		list = append(list, f)
	}
	return list, rows.Err()
}

func getByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*Record, error) {
	var rec Record
	err := pool.QueryRow(ctx, `
		SELECT id, requester_id, addressee_id, status::text, created_at, updated_at
		FROM friendships WHERE id = $1
	`, id).Scan(&rec.ID, &rec.RequesterID, &rec.AddresseeID, &rec.Status, &rec.CreatedAt, &rec.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFriendshipNotFound
		}
		return nil, err
	}
	return &rec, nil
}

func insert(ctx context.Context, pool *pgxpool.Pool, requesterID, addresseeID uuid.UUID) (*Record, error) {
	var rec Record
	err := pool.QueryRow(ctx, `
		INSERT INTO friendships (requester_id, addressee_id, status)
		VALUES ($1, $2, 'pending')
		RETURNING id, requester_id, addressee_id, status::text, created_at, updated_at
	`, requesterID, addresseeID).Scan(
		&rec.ID, &rec.RequesterID, &rec.AddresseeID, &rec.Status, &rec.CreatedAt, &rec.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create friendship: %w", err)
	}
	return &rec, nil
}

func updateRequest(ctx context.Context, pool *pgxpool.Pool, id, requesterID, addresseeID uuid.UUID) (*Record, error) {
	var rec Record
	err := pool.QueryRow(ctx, `
		UPDATE friendships
		SET requester_id = $2, addressee_id = $3, status = 'pending', updated_at = NOW()
		WHERE id = $1
		RETURNING id, requester_id, addressee_id, status::text, created_at, updated_at
	`, id, requesterID, addresseeID).Scan(
		&rec.ID, &rec.RequesterID, &rec.AddresseeID, &rec.Status, &rec.CreatedAt, &rec.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update friendship: %w", err)
	}
	return &rec, nil
}

func setStatus(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, status string) (*Record, error) {
	var rec Record
	err := pool.QueryRow(ctx, `
		UPDATE friendships
		SET status = $2::friendship_status, updated_at = NOW()
		WHERE id = $1
		RETURNING id, requester_id, addressee_id, status::text, created_at, updated_at
	`, id, status).Scan(
		&rec.ID, &rec.RequesterID, &rec.AddresseeID, &rec.Status, &rec.CreatedAt, &rec.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFriendshipNotFound
		}
		return nil, err
	}
	return &rec, nil
}
