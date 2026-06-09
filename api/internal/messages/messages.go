package messages

import (
	"bytes"
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
	ErrUserNotFound         = errors.New("user not found")
	ErrConversationNotFound = errors.New("conversation not found")
	ErrNotAuthorized        = errors.New("not authorized")
	ErrCannotMessage        = errors.New("can only message snack mates")
	ErrEmptyBody            = errors.New("message body required")
	ErrEmptySubject         = errors.New("message subject required")
	ErrBodyTooLong          = errors.New("message too long")
	ErrSubjectTooLong       = errors.New("message subject too long")
)

const maxBodyLen = 4000
const maxSubjectLen = 200

type Record struct {
	ID        uuid.UUID
	UserAID   uuid.UUID
	UserBID   uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MessageRecord struct {
	ID             uuid.UUID
	ConversationID uuid.UUID
	SenderID       uuid.UUID
	Body           string
	CreatedAt      time.Time
	ReadAt         *time.Time
}

func CanMessage(ctx context.Context, pool *pgxpool.Pool, userA, userB uuid.UUID) bool {
	if userA == userB {
		return false
	}
	var exists bool
	_ = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM friendships f
			WHERE f.status = 'accepted'
			  AND (($1 = f.requester_id AND $2 = f.addressee_id)
			    OR ($1 = f.addressee_id AND $2 = f.requester_id))
		) OR EXISTS (
			SELECT 1 FROM snack_matches sm
			WHERE sm.status IN ('pending', 'active')
			  AND $1 IN (sm.user_a_id, sm.user_b_id)
			  AND $2 IN (sm.user_a_id, sm.user_b_id)
		)
	`, userA, userB).Scan(&exists)
	return exists
}

func orderedPair(a, b uuid.UUID) (uuid.UUID, uuid.UUID) {
	if bytes.Compare(a[:], b[:]) < 0 {
		return a, b
	}
	return b, a
}

func participantIn(conv Record, userID uuid.UUID) bool {
	return conv.UserAID == userID || conv.UserBID == userID
}

func otherParticipant(conv Record, userID uuid.UUID) uuid.UUID {
	if conv.UserAID == userID {
		return conv.UserBID
	}
	return conv.UserAID
}

func OtherUser(ctx context.Context, pool *pgxpool.Pool, conv Record, viewerID uuid.UUID) (*models.PublicUser, error) {
	otherID := otherParticipant(conv, viewerID)
	var user models.PublicUser
	var avatarKey, avatarURL *string
	err := pool.QueryRow(ctx, `
		SELECT id, username, display_name, COALESCE(bio, ''), COALESCE(country, ''), avatar_key, avatar_url
		FROM users
		WHERE id = $1
		  AND email_verified = TRUE
		  AND deactivated_at IS NULL
	`, otherID).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.Bio, &user.Country,
		&avatarKey, &avatarURL,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if avatarURL != nil && *avatarURL != "" {
		user.AvatarURL = avatarURL
	}
	return &user, nil
}

func GetByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*Record, error) {
	var rec Record
	err := pool.QueryRow(ctx, `
		SELECT id, user_a_id, user_b_id, created_at, updated_at
		FROM conversations WHERE id = $1
	`, id).Scan(&rec.ID, &rec.UserAID, &rec.UserBID, &rec.CreatedAt, &rec.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrConversationNotFound
		}
		return nil, err
	}
	return &rec, nil
}

func FindOrCreate(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, username string) (*Record, error) {
	target, err := auth.GetUserByUsername(ctx, pool, username)
	if err != nil || !target.EmailVerified || target.DeactivatedAt != nil {
		return nil, ErrUserNotFound
	}
	if target.ID == userID {
		return nil, ErrCannotMessage
	}
	if !CanMessage(ctx, pool, userID, target.ID) {
		return nil, ErrCannotMessage
	}
	return findOrCreateBetween(ctx, pool, userID, target.ID)
}

func findOrCreateBetween(ctx context.Context, pool *pgxpool.Pool, userA, userB uuid.UUID) (*Record, error) {
	low, high := orderedPair(userA, userB)

	var rec Record
	err := pool.QueryRow(ctx, `
		INSERT INTO conversations (user_a_id, user_b_id)
		VALUES ($1, $2)
		ON CONFLICT (user_a_id, user_b_id) DO UPDATE SET updated_at = conversations.updated_at
		RETURNING id, user_a_id, user_b_id, created_at, updated_at
	`, low, high).Scan(&rec.ID, &rec.UserAID, &rec.UserBID, &rec.CreatedAt, &rec.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("find or create conversation: %w", err)
	}
	return &rec, nil
}

func ListConversations(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) ([]models.Conversation, error) {
	rows, err := pool.Query(ctx, `
		SELECT c.id, c.user_a_id, c.user_b_id, c.created_at, c.updated_at,
		       u.id, u.username, u.display_name, COALESCE(u.bio, ''), COALESCE(u.country, ''),
		       u.avatar_key, u.avatar_url,
		       lm.id, lm.sender_id, lm.subject, lm.body, lm.created_at, lm.read_at,
		       (
		           SELECT COUNT(*)::int FROM messages m
		           WHERE m.conversation_id = c.id
		             AND m.sender_id != $1
		             AND m.read_at IS NULL
		       )
		FROM conversations c
		JOIN users u ON u.id = CASE
			WHEN c.user_a_id = $1 THEN c.user_b_id
			ELSE c.user_a_id
		END
		LEFT JOIN LATERAL (
			SELECT id, sender_id, subject, body, created_at, read_at
			FROM messages
			WHERE conversation_id = c.id
			ORDER BY created_at DESC
			LIMIT 1
		) lm ON TRUE
		WHERE c.user_a_id = $1 OR c.user_b_id = $1
		ORDER BY c.updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]models.Conversation, 0)
	for rows.Next() {
		var conv models.Conversation
		var user models.PublicUser
		var avatarKey, avatarURL *string
		var lastID, lastSenderID *uuid.UUID
		var lastSubject, lastBody *string
		var lastCreatedAt *time.Time
		var lastReadAt *time.Time
		if err := rows.Scan(
			&conv.ID, &conv.UserAID, &conv.UserBID, &conv.CreatedAt, &conv.UpdatedAt,
			&user.ID, &user.Username, &user.DisplayName, &user.Bio, &user.Country,
			&avatarKey, &avatarURL,
			&lastID, &lastSenderID, &lastSubject, &lastBody, &lastCreatedAt, &lastReadAt,
			&conv.UnreadCount,
		); err != nil {
			return nil, err
		}
		if avatarURL != nil && *avatarURL != "" {
			user.AvatarURL = avatarURL
		}
		conv.OtherUser = &user
		if lastID != nil {
			conv.LastMessage = &models.Message{
				ID:             *lastID,
				ConversationID: conv.ID,
				SenderID:       *lastSenderID,
				Subject:        *lastSubject,
				Body:           *lastBody,
				CreatedAt:      *lastCreatedAt,
				ReadAt:         lastReadAt,
			}
		}
		list = append(list, conv)
	}
	return list, rows.Err()
}

func ListMessages(ctx context.Context, pool *pgxpool.Pool, conversationID, userID uuid.UUID, limit int) ([]models.Message, error) {
	conv, err := GetByID(ctx, pool, conversationID)
	if err != nil {
		return nil, err
	}
	if !participantIn(*conv, userID) {
		return nil, ErrNotAuthorized
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := pool.Query(ctx, `
		SELECT id, conversation_id, sender_id, subject, body, created_at, read_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
		LIMIT $2
	`, conversationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]models.Message, 0)
	for rows.Next() {
		var msg models.Message
		if err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Subject, &msg.Body, &msg.CreatedAt, &msg.ReadAt); err != nil {
			return nil, err
		}
		list = append(list, msg)
	}
	return list, rows.Err()
}

func Send(ctx context.Context, pool *pgxpool.Pool, conversationID, senderID uuid.UUID, subject, body string) (*models.Message, *Record, error) {
	subject = trimBody(subject)
	body = trimBody(body)
	if subject == "" {
		return nil, nil, ErrEmptySubject
	}
	if body == "" {
		return nil, nil, ErrEmptyBody
	}
	if len(subject) > maxSubjectLen {
		return nil, nil, ErrSubjectTooLong
	}
	if len(body) > maxBodyLen {
		return nil, nil, ErrBodyTooLong
	}

	conv, err := GetByID(ctx, pool, conversationID)
	if err != nil {
		return nil, nil, err
	}
	if !participantIn(*conv, senderID) {
		return nil, nil, ErrNotAuthorized
	}
	otherID := otherParticipant(*conv, senderID)
	if !CanMessage(ctx, pool, senderID, otherID) {
		return nil, nil, ErrCannotMessage
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	var msg models.Message
	err = tx.QueryRow(ctx, `
		INSERT INTO messages (conversation_id, sender_id, subject, body)
		VALUES ($1, $2, $3, $4)
		RETURNING id, conversation_id, sender_id, subject, body, created_at, read_at
	`, conversationID, senderID, subject, body).Scan(
		&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Subject, &msg.Body, &msg.CreatedAt, &msg.ReadAt,
	)
	if err != nil {
		return nil, nil, err
	}

	err = tx.QueryRow(ctx, `
		UPDATE conversations SET updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_a_id, user_b_id, created_at, updated_at
	`, conversationID).Scan(&conv.ID, &conv.UserAID, &conv.UserBID, &conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}
	return &msg, conv, nil
}

func MarkRead(ctx context.Context, pool *pgxpool.Pool, conversationID, userID uuid.UUID) error {
	conv, err := GetByID(ctx, pool, conversationID)
	if err != nil {
		return err
	}
	if !participantIn(*conv, userID) {
		return ErrNotAuthorized
	}

	_, err = pool.Exec(ctx, `
		UPDATE messages
		SET read_at = NOW()
		WHERE conversation_id = $1
		  AND sender_id != $2
		  AND read_at IS NULL
	`, conversationID, userID)
	return err
}

func UnreadTotal(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) (int, error) {
	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM messages m
		JOIN conversations c ON c.id = m.conversation_id
		WHERE (c.user_a_id = $1 OR c.user_b_id = $1)
		  AND m.sender_id != $1
		  AND m.read_at IS NULL
	`, userID).Scan(&count)
	return count, err
}

func trimBody(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
