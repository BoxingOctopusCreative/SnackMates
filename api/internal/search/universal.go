package search

import (
	"context"
	"strings"
	"sync"

	"github.com/boxingoctopus/snackmates/api/internal/snacksearch"
	"github.com/boxingoctopus/snackmates/api/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultUserLimit     = 8
	defaultWishlistLimit = 10
	defaultProductLimit  = 12
)

type UniversalResult struct {
	Query         string                   `json:"query"`
	SearchTerms   string                   `json:"search_terms"`
	AIAssisted    bool                     `json:"ai_assisted"`
	Users         []PublicUserHit          `json:"users"`
	WishlistItems []WishlistItemHit        `json:"wishlist_items"`
	Products      []snacksearch.ProductHit `json:"products"`
}

type PublicUserHit struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	DisplayName string  `json:"display_name"`
	Bio         string  `json:"bio"`
	Country     string  `json:"country"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

type WishlistItemHit struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	Brand         string  `json:"brand"`
	Notes         string  `json:"notes"`
	Score         float64 `json:"score,omitempty"`
	UserID        string  `json:"user_id"`
	UserName      string  `json:"user_name"`
	Username      string  `json:"username"`
	WishlistSlug  string  `json:"wishlist_slug"`
	WishlistTitle string  `json:"wishlist_title"`
}

func UniversalSearch(
	ctx context.Context,
	pool *pgxpool.Pool,
	meili *Client,
	products *snacksearch.Service,
	storage *storage.Client,
	query string,
) (UniversalResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return UniversalResult{
			Users:         []PublicUserHit{},
			WishlistItems: []WishlistItemHit{},
			Products:      []snacksearch.ProductHit{},
		}, nil
	}

	var (
		wg           sync.WaitGroup
		users        []UserHit
		wishlistHits []WishlistItemHit
		productResp  snacksearch.Response
		userErr      error
		wishlistErr  error
		productErr   error
	)

	wg.Add(3)
	go func() {
		defer wg.Done()
		users, userErr = SearchUsers(ctx, pool, query, defaultUserLimit)
	}()
	go func() {
		defer wg.Done()
		wishlistHits, wishlistErr = meili.SearchWishlistItems(ctx, query, defaultWishlistLimit)
	}()
	go func() {
		defer wg.Done()
		productResp, productErr = products.Search(ctx, query, defaultProductLimit)
	}()
	wg.Wait()

	result := UniversalResult{
		Query:         query,
		SearchTerms:   query,
		Users:         []PublicUserHit{},
		WishlistItems: []WishlistItemHit{},
		Products:      []snacksearch.ProductHit{},
	}

	if userErr == nil {
		result.Users = publicUserHits(ctx, storage, users)
	}
	if wishlistErr == nil {
		result.WishlistItems = wishlistHits
	}
	if productErr == nil {
		result.SearchTerms = productResp.SearchTerms
		result.AIAssisted = productResp.AIAssisted
		if productResp.Results != nil {
			result.Products = productResp.Results
		}
	}

	return result, nil
}

func publicUserHits(ctx context.Context, s *storage.Client, users []UserHit) []PublicUserHit {
	hits := make([]PublicUserHit, 0, len(users))
	for _, u := range users {
		hit := PublicUserHit{
			ID:          u.ID.String(),
			Username:    u.Username,
			DisplayName: u.DisplayName,
			Bio:         u.Bio,
			Country:     u.Country,
		}
		if s != nil {
			if avatarURL, err := s.ResolveObjectURL(ctx, u.AvatarURL, u.AvatarKey); err == nil && avatarURL != "" {
				hit.AvatarURL = &avatarURL
			}
		} else if u.AvatarURL != nil && *u.AvatarURL != "" {
			hit.AvatarURL = u.AvatarURL
		}
		hits = append(hits, hit)
	}
	return hits
}
