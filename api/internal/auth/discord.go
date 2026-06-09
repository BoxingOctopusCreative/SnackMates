package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/boxingoctopus/snackmates/api/internal/config"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
)

const (
	oauthPurposeLogin   = "login"
	oauthPurposeConnect = "connect:"
)

var discordEndpoint = oauth2.Endpoint{
	AuthURL:  "https://discord.com/api/oauth2/authorize",
	TokenURL: "https://discord.com/api/oauth2/token",
}

type DiscordService struct {
	cfg   config.Config
	oauth *oauth2.Config
}

func NewDiscordService(cfg config.Config) *DiscordService {
	return &DiscordService{
		cfg: cfg,
		oauth: &oauth2.Config{
			ClientID:     cfg.DiscordClientID,
			ClientSecret: cfg.DiscordSecret,
			RedirectURL:  cfg.DiscordRedirect,
			Endpoint:     discordEndpoint,
			Scopes:       []string{"identify", "email"},
		},
	}
}

func (d *DiscordService) Enabled() bool {
	return d.cfg.DiscordClientID != "" && d.cfg.DiscordSecret != ""
}

func (d *DiscordService) AuthURL(state string) string {
	return d.oauth.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// HandleLoginCallback signs in or registers a user via Discord OAuth.
func (d *DiscordService) HandleLoginCallback(ctx context.Context, pool *pgxpool.Pool, code string) (uuid.UUID, string, error) {
	du, profile, err := d.fetchProfile(ctx, code)
	if err != nil {
		return uuid.Nil, "", err
	}
	if du.Email == "" {
		return uuid.Nil, "", fmt.Errorf("discord account has no email")
	}

	userID, err := upsertDiscordUser(ctx, pool, du, profile)
	if err != nil {
		return uuid.Nil, "", err
	}

	var deactivatedAt interface{}
	if err := pool.QueryRow(ctx, `SELECT deactivated_at FROM users WHERE id = $1`, userID).Scan(&deactivatedAt); err != nil {
		return uuid.Nil, "", err
	}
	if deactivatedAt != nil {
		return uuid.Nil, "", ErrAccountDeactivated
	}

	sessionToken, err := CreateSession(ctx, pool, userID)
	if err != nil {
		return uuid.Nil, "", err
	}
	return userID, sessionToken, nil
}

// LinkAccount connects Discord to an existing SnackMates account and syncs profile fields.
func (d *DiscordService) LinkAccount(ctx context.Context, pool *pgxpool.Pool, code string, userID uuid.UUID) error {
	du, profile, err := d.fetchProfile(ctx, code)
	if err != nil {
		return err
	}

	var linkedUserID uuid.UUID
	err = pool.QueryRow(ctx, `SELECT id FROM users WHERE discord_id = $1`, du.ID).Scan(&linkedUserID)
	if err == nil && linkedUserID != userID {
		return fmt.Errorf("discord account is already linked to another user")
	}
	if err != nil && err != pgx.ErrNoRows {
		return err
	}

	var currentDiscordID *string
	err = pool.QueryRow(ctx, `SELECT discord_id FROM users WHERE id = $1`, userID).Scan(&currentDiscordID)
	if err != nil {
		return err
	}
	if currentDiscordID != nil && *currentDiscordID != "" && *currentDiscordID != du.ID {
		return fmt.Errorf("account is already linked to a different discord profile")
	}

	return syncDiscordProfile(ctx, pool, userID, du.ID, profile)
}

func (d *DiscordService) fetchProfile(ctx context.Context, code string) (discordUser, DiscordProfile, error) {
	token, err := d.oauth.Exchange(ctx, code)
	if err != nil {
		return discordUser{}, DiscordProfile{}, fmt.Errorf("exchange code: %w", err)
	}

	du, err := d.fetchDiscordUser(ctx, token.AccessToken)
	if err != nil {
		return discordUser{}, DiscordProfile{}, err
	}
	return du, ProfileFromDiscordUser(du), nil
}

func (d *DiscordService) fetchDiscordUser(ctx context.Context, accessToken string) (discordUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://discord.com/api/users/@me", nil)
	if err != nil {
		return discordUser{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return discordUser{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return discordUser{}, fmt.Errorf("discord user request failed: %s", resp.Status)
	}

	var du discordUser
	if err := json.NewDecoder(resp.Body).Decode(&du); err != nil {
		return discordUser{}, err
	}
	return du, nil
}

func upsertDiscordUser(ctx context.Context, pool *pgxpool.Pool, du discordUser, profile DiscordProfile) (uuid.UUID, error) {
	var userID uuid.UUID
	err := pool.QueryRow(ctx, `SELECT id FROM users WHERE discord_id = $1`, du.ID).Scan(&userID)
	if err == nil {
		if err := syncDiscordProfile(ctx, pool, userID, du.ID, profile); err != nil {
			return uuid.Nil, err
		}
		return userID, nil
	}
	if err != pgx.ErrNoRows {
		return uuid.Nil, err
	}

	username, err := allocateUsername(ctx, pool, profile.DisplayName)
	if err != nil {
		return uuid.Nil, err
	}
	err = pool.QueryRow(ctx, `
		INSERT INTO users (email, display_name, bio, discord_id, email_verified, avatar_url, discord_avatar_hash, username)
		VALUES ($1, $2, $3, $4, TRUE, $5, $6, $7)
		ON CONFLICT (email) DO UPDATE SET
			discord_id = EXCLUDED.discord_id,
			display_name = EXCLUDED.display_name,
			bio = COALESCE(NULLIF(EXCLUDED.bio, ''), users.bio),
			avatar_url = EXCLUDED.avatar_url,
			discord_avatar_hash = EXCLUDED.discord_avatar_hash,
			avatar_key = NULL
		RETURNING id
	`, du.Email, profile.DisplayName, profile.Bio, du.ID, profile.AvatarURL, profile.AvatarHash, username).Scan(&userID)
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

func syncDiscordProfile(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, discordID string, profile DiscordProfile) error {
	_, err := pool.Exec(ctx, `
		UPDATE users
		SET discord_id = $2,
		    display_name = $3,
		    bio = CASE WHEN NULLIF(TRIM($4), '') IS NOT NULL THEN $4 ELSE bio END,
		    avatar_url = $5,
		    discord_avatar_hash = $6,
		    avatar_key = NULL
		WHERE id = $1
	`, userID, discordID, profile.DisplayName, profile.Bio, profile.AvatarURL, profile.AvatarHash)
	return err
}

func StoreOAuthState(cache interface{ Set(context.Context, string, string, time.Duration) error }, state, purpose string) error {
	return cache.Set(context.Background(), "oauth:"+state, purpose, 10*time.Minute)
}

func ConsumeOAuthState(cache interface {
	Get(context.Context, string) (string, error)
	Delete(context.Context, ...string) error
}, state string) (string, error) {
	key := "oauth:" + state
	purpose, err := cache.Get(context.Background(), key)
	if err != nil {
		return "", fmt.Errorf("invalid oauth state")
	}
	if err := cache.Delete(context.Background(), key); err != nil {
		return "", err
	}
	return purpose, nil
}

func OAuthPurposeLogin() string {
	return oauthPurposeLogin
}

func ConnectOAuthPurpose(userID uuid.UUID) string {
	return oauthPurposeConnect + userID.String()
}

func IsConnectOAuthPurpose(purpose string) bool {
	return strings.HasPrefix(purpose, oauthPurposeConnect)
}

func UserIDFromConnectPurpose(purpose string) (uuid.UUID, error) {
	raw := strings.TrimPrefix(purpose, oauthPurposeConnect)
	return uuid.Parse(raw)
}
