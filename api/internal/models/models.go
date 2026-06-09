package models

import (
	"time"

	"github.com/google/uuid"
)

type PublicUser struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Bio         string    `json:"bio"`
	Country     string    `json:"country"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
	BannerURL   *string   `json:"banner_url,omitempty"`
}

type FriendshipView struct {
	ID     *uuid.UUID `json:"id,omitempty"`
	Status string     `json:"status"`
}

type Friendship struct {
	ID          uuid.UUID  `json:"id"`
	RequesterID uuid.UUID  `json:"requester_id"`
	AddresseeID uuid.UUID  `json:"addressee_id"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	User        *PublicUser `json:"user,omitempty"`
}

type UserProfile struct {
	User       PublicUser      `json:"user"`
	Wishlists  []Wishlist      `json:"wishlists"`
	Friendship *FriendshipView `json:"friendship,omitempty"`
}

type User struct {
	ID            uuid.UUID `json:"id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	DisplayName   string    `json:"display_name"`
	Bio           string    `json:"bio"`
	Country       string    `json:"country"`
	AvatarKey     *string   `json:"avatar_key,omitempty"`
	AvatarURL     *string   `json:"avatar_url,omitempty"`
	BannerKey     *string   `json:"banner_key,omitempty"`
	BannerURL     *string   `json:"banner_url,omitempty"`
	DiscordID     *string   `json:"discord_id,omitempty"`
	TOTPEnabled   bool      `json:"totp_enabled"`
	HasWebAuthn   bool      `json:"has_webauthn"`
	CreatedAt     time.Time `json:"created_at"`
}

type Wishlist struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	IsPublic    bool      `json:"is_public"`
	ItemCount   int       `json:"item_count"`
	BannerURL   *string   `json:"banner_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type FriendWishlist struct {
	Wishlist
	Owner PublicUser `json:"owner"`
}

type SnaggedBy struct {
	ID             uuid.UUID `json:"id"`
	DisplayName    string    `json:"display_name"`
	DeliveryMethod string    `json:"delivery_method"`
	TrackingNumber *string   `json:"tracking_number,omitempty"`
}

type WishlistItem struct {
	ID         uuid.UUID  `json:"id"`
	WishlistID uuid.UUID  `json:"wishlist_id"`
	Name       string     `json:"name"`
	Type       string     `json:"type"`
	Brand      string     `json:"brand"`
	Notes      string     `json:"notes"`
	ImageURL   string     `json:"image_url,omitempty"`
	SnaggedBy  *SnaggedBy `json:"snagged_by,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type WishlistDetail struct {
	Wishlist      Wishlist       `json:"wishlist"`
	Items         []WishlistItem `json:"items"`
	ViewerCanSnag bool           `json:"viewer_can_snag"`
}

type SnackMatch struct {
	ID        uuid.UUID `json:"id"`
	UserAID   uuid.UUID `json:"user_a_id"`
	UserBID   uuid.UUID `json:"user_b_id"`
	Status    string    `json:"status"`
	MatchedAt time.Time `json:"matched_at"`
	Mate      *User     `json:"mate,omitempty"`
}

type Message struct {
	ID             uuid.UUID  `json:"id"`
	ConversationID uuid.UUID  `json:"conversation_id"`
	SenderID       uuid.UUID  `json:"sender_id"`
	Subject        string     `json:"subject"`
	Body           string     `json:"body"`
	CreatedAt      time.Time  `json:"created_at"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
}

type Chat struct {
	ID          uuid.UUID   `json:"id"`
	UserAID     uuid.UUID   `json:"user_a_id"`
	UserBID     uuid.UUID   `json:"user_b_id"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	OtherUser   *PublicUser `json:"other_user,omitempty"`
	LastMessage *ChatMessage `json:"last_message,omitempty"`
	UnreadCount int         `json:"unread_count"`
}

type ChatMessage struct {
	ID        uuid.UUID  `json:"id"`
	ChatID    uuid.UUID  `json:"chat_id"`
	SenderID  uuid.UUID  `json:"sender_id"`
	Body      string     `json:"body"`
	CreatedAt time.Time  `json:"created_at"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
}

type Conversation struct {
	ID          uuid.UUID   `json:"id"`
	UserAID     uuid.UUID   `json:"user_a_id"`
	UserBID     uuid.UUID   `json:"user_b_id"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	OtherUser   *PublicUser `json:"other_user,omitempty"`
	LastMessage *Message    `json:"last_message,omitempty"`
	UnreadCount int         `json:"unread_count"`
}

type SearchHit struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Brand    string  `json:"brand"`
	Notes    string  `json:"notes"`
	Score    float64 `json:"score"`
	UserID   string  `json:"user_id"`
	UserName string  `json:"user_name"`
}
