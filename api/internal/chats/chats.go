package chats

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/boxingoctopus/snackmates/api/internal/auth"
	"github.com/boxingoctopus/snackmates/api/internal/messages"
	"github.com/boxingoctopus/snackmates/api/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrChatNotFound    = errors.New("chat not found")
	ErrNotAuthorized   = errors.New("not authorized")
	ErrCannotChat      = errors.New("can only chat with snack mates")
	ErrEmptyBody       = errors.New("message body required")
	ErrBodyTooLong     = errors.New("message too long")
)

const maxBodyLen = 500

type Record struct {
	ID        uuid.UUID
	UserAID   uuid.UUID
	UserBID   uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

func orderedPair(a, b uuid.UUID) (uuid.UUID, uuid.UUID) {
	if bytes.Compare(a[:], b[:]) < 0 {
		return a, b
	}
	return b, a
}

func participantIn(chat Record, userID uuid.UUID) bool {
	return chat.UserAID == userID || chat.UserBID == userID
}

func otherParticipant(chat Record, userID uuid.UUID) uuid.UUID {
	if chat.UserAID == userID {
		return chat.UserBID
	}
	return chat.UserAID
}

func GetByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*Record, error) {
	var rec Record
	err := pool.QueryRow(ctx, `
		SELECT id, user_a_id, user_b_id, created_at, updated_at
		FROM chats WHERE id = $1
	`, id).Scan(&rec.ID, &rec.UserAID, &rec.UserBID, &rec.CreatedAt, &rec.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrChatNotFound
		}
		return nil, err
	}
	return &rec, nil
}

func OtherUser(ctx context.Context, pool *pgxpool.Pool, chat Record, viewerID uuid.UUID) (*models.PublicUser, error) {
	otherID := otherParticipant(chat, viewerID)
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

func FindOrCreate(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, username string) (*Record, error) {
	target, err := auth.GetUserByUsername(ctx, pool, username)
	if err != nil || !target.EmailVerified || target.DeactivatedAt != nil {
		return nil, ErrUserNotFound
	}
	if target.ID == userID {
		return nil, ErrCannotChat
	}
	if !messages.CanMessage(ctx, pool, userID, target.ID) {
		return nil, ErrCannotChat
	}
	return findOrCreateBetween(ctx, pool, userID, target.ID)
}

func findOrCreateBetween(ctx context.Context, pool *pgxpool.Pool, userA, userB uuid.UUID) (*Record, error) {
	low, high := orderedPair(userA, userB)

	var rec Record
	err := pool.QueryRow(ctx, `
		INSERT INTO chats (user_a_id, user_b_id)
		VALUES ($1, $2)
		ON CONFLICT (user_a_id, user_b_id) DO UPDATE SET updated_at = chats.updated_at
		RETURNING id, user_a_id, user_b_id, created_at, updated_at
	`, low, high).Scan(&rec.ID, &rec.UserAID, &rec.UserBID, &rec.CreatedAt, &rec.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("find or create chat: %w", err)
	}
	return &rec, nil
}

func ListChats(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) ([]models.Chat, error) {
	rows, err := pool.Query(ctx, `
		SELECT c.id, c.user_a_id, c.user_b_id, c.created_at, c.updated_at,
		       u.id, u.username, u.display_name, COALESCE(u.bio, ''), COALESCE(u.country, ''),
		       u.avatar_key, u.avatar_url,
		       lm.id, lm.sender_id, lm.body, lm.created_at, lm.read_at,
		       (
		           SELECT COUNT(*)::int FROM chat_messages m
		           WHERE m.chat_id = c.id
		             AND m.sender_id != $1
		             AND m.read_at IS NULL
		       )
		FROM chats c
		JOIN users u ON u.id = CASE
			WHEN c.user_a_id = $1 THEN c.user_b_id
			ELSE c.user_a_id
		END
		LEFT JOIN LATERAL (
			SELECT id, sender_id, body, created_at, read_at
			FROM chat_messages
			WHERE chat_id = c.id
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

	list := make([]models.Chat, 0)
	for rows.Next() {
		var chat models.Chat
		var user models.PublicUser
		var avatarKey, avatarURL *string
		var lastID, lastSenderID *uuid.UUID
		var lastBody *string
		var lastCreatedAt *time.Time
		var lastReadAt *time.Time
		if err := rows.Scan(
			&chat.ID, &chat.UserAID, &chat.UserBID, &chat.CreatedAt, &chat.UpdatedAt,
			&user.ID, &user.Username, &user.DisplayName, &user.Bio, &user.Country,
			&avatarKey, &avatarURL,
			&lastID, &lastSenderID, &lastBody, &lastCreatedAt, &lastReadAt,
			&chat.UnreadCount,
		); err != nil {
			return nil, err
		}
		if avatarURL != nil && *avatarURL != "" {
			user.AvatarURL = avatarURL
		}
		chat.OtherUser = &user
		if lastID != nil {
			chat.LastMessage = &models.ChatMessage{
				ID:        *lastID,
				ChatID:    chat.ID,
				SenderID:  *lastSenderID,
				Body:      *lastBody,
				CreatedAt: *lastCreatedAt,
				ReadAt:    lastReadAt,
			}
		}
		list = append(list, chat)
	}
	return list, rows.Err()
}

func ListMessages(ctx context.Context, pool *pgxpool.Pool, chatID, userID uuid.UUID, limit int) ([]models.ChatMessage, error) {
	chat, err := GetByID(ctx, pool, chatID)
	if err != nil {
		return nil, err
	}
	if !participantIn(*chat, userID) {
		return nil, ErrNotAuthorized
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := pool.Query(ctx, `
		SELECT id, chat_id, sender_id, body, created_at, read_at
		FROM chat_messages
		WHERE chat_id = $1
		ORDER BY created_at ASC
		LIMIT $2
	`, chatID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]models.ChatMessage, 0)
	for rows.Next() {
		var msg models.ChatMessage
		if err := rows.Scan(&msg.ID, &msg.ChatID, &msg.SenderID, &msg.Body, &msg.CreatedAt, &msg.ReadAt); err != nil {
			return nil, err
		}
		list = append(list, msg)
	}
	return list, rows.Err()
}

func Send(ctx context.Context, pool *pgxpool.Pool, chatID, senderID uuid.UUID, body string) (*models.ChatMessage, *Record, error) {
	body = trimBody(body)
	if body == "" {
		return nil, nil, ErrEmptyBody
	}
	if len(body) > maxBodyLen {
		return nil, nil, ErrBodyTooLong
	}

	chat, err := GetByID(ctx, pool, chatID)
	if err != nil {
		return nil, nil, err
	}
	if !participantIn(*chat, senderID) {
		return nil, nil, ErrNotAuthorized
	}
	otherID := otherParticipant(*chat, senderID)
	if !messages.CanMessage(ctx, pool, senderID, otherID) {
		return nil, nil, ErrCannotChat
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	var msg models.ChatMessage
	err = tx.QueryRow(ctx, `
		INSERT INTO chat_messages (chat_id, sender_id, body)
		VALUES ($1, $2, $3)
		RETURNING id, chat_id, sender_id, body, created_at, read_at
	`, chatID, senderID, body).Scan(
		&msg.ID, &msg.ChatID, &msg.SenderID, &msg.Body, &msg.CreatedAt, &msg.ReadAt,
	)
	if err != nil {
		return nil, nil, err
	}

	err = tx.QueryRow(ctx, `
		UPDATE chats SET updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_a_id, user_b_id, created_at, updated_at
	`, chatID).Scan(&chat.ID, &chat.UserAID, &chat.UserBID, &chat.CreatedAt, &chat.UpdatedAt)
	if err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}
	return &msg, chat, nil
}

func MarkRead(ctx context.Context, pool *pgxpool.Pool, chatID, userID uuid.UUID) error {
	chat, err := GetByID(ctx, pool, chatID)
	if err != nil {
		return err
	}
	if !participantIn(*chat, userID) {
		return ErrNotAuthorized
	}

	_, err = pool.Exec(ctx, `
		UPDATE chat_messages
		SET read_at = NOW()
		WHERE chat_id = $1
		  AND sender_id != $2
		  AND read_at IS NULL
	`, chatID, userID)
	return err
}

func UnreadTotal(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) (int, error) {
	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM chat_messages m
		JOIN chats c ON c.id = m.chat_id
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
