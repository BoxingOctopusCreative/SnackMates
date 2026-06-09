package seed

import (
	"context"
	"fmt"
	"log"

	"github.com/boxingoctopus/snackmates/api/internal/auth"
	"github.com/boxingoctopus/snackmates/api/internal/search"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const DefaultPassword = "snackmates123"

// BrunoFaviconURL is the default avatar for the automated bot test user.
const BrunoFaviconURL = "https://assets.snackmates.food/brand/logokit_favicon.png"

type Item struct {
	Name  string
	Type  string
	Brand string
	Notes string
}

type Wishlist struct {
	Slug        string
	Title       string
	Description string
	IsPublic    bool
	Items       []Item
}

type User struct {
	Email       string
	Username    string
	DisplayName string
	Bio         string
	Country     string
	AvatarURL   string
	Wishlists   []Wishlist
}

var Users = []User{
	{
		Email:       "alice@snackmates.local",
		Username:    "alice",
		DisplayName: "Alice Chen",
		Bio:         "Always hunting for sour gummies and Japanese kit kats.",
		Country:     "US",
		Wishlists: []Wishlist{
			{
				Slug:        "alice-favorites",
				Title:       "Alice's Favorites",
				Description: "Snacks I want to try from pen pals abroad.",
				IsPublic:    true,
				Items: []Item{
					{Name: "Hi-Chew", Type: "Candy", Brand: "Morinaga", Notes: "Grape or mango flavors"},
					{Name: "Pocky", Type: "Baked Goods", Brand: "Glico", Notes: "Matcha or strawberry"},
					{Name: "Lay's Magic Masala", Type: "Chips/Crackers", Brand: "Lay's", Notes: "India exclusive"},
				},
			},
		},
	},
	{
		Email:       "bruno@snackmates.local",
		Username:    "bruno",
		DisplayName: "Bruno Müller",
		Bio:         "Chocolate collector from Berlin.",
		Country:     "DE",
		AvatarURL:   BrunoFaviconURL,
		Wishlists: []Wishlist{
			{
				Slug:        "bruno-chocolate",
				Title:       "Bruno's Chocolate Box",
				Description: "European chocolates and biscuits.",
				IsPublic:    true,
				Items: []Item{
					{Name: "Milka Oreo", Type: "Candy", Brand: "Milka", Notes: ""},
					{Name: "Leibniz Butterkeks", Type: "Baked Goods", Brand: "Bahlsen", Notes: "Classic butter biscuits"},
					{Name: "Club-Mate", Type: "Beverages", Brand: "Club-Mate", Notes: "The original"},
				},
			},
		},
	},
	{
		Email:       "carmen@snackmates.local",
		Username:    "carmen",
		DisplayName: "Carmen Dubois",
		Bio:         "Maple everything, please.",
		Country:     "CA",
		Wishlists: []Wishlist{
			{
				Slug:        "carmen-maple",
				Title:       "Maple & More",
				Description: "Canadian classics to share.",
				IsPublic:    true,
				Items: []Item{
					{Name: "Coffee Crisp", Type: "Candy", Brand: "Nestlé", Notes: ""},
					{Name: "Ketchup Chips", Type: "Chips/Crackers", Brand: "Lay's", Notes: "Canadian style"},
					{Name: "Maple Cookies", Type: "Baked Goods", Brand: "President's Choice", Notes: ""},
				},
			},
			{
				Slug:        "carmen-private-stash",
				Title:       "Private Stash",
				Description: "Only for close snack mates.",
				IsPublic:    false,
				Items: []Item{
					{Name: "Smarties", Type: "Candy", Brand: "Nestlé", Notes: "The Canadian kind"},
				},
			},
		},
	},
	{
		Email:       "diego@snackmates.local",
		Username:    "diego",
		DisplayName: "Diego Rivera",
		Bio:         "Spicy snacks and tamarind candy from Mexico City.",
		Country:     "MX",
		Wishlists: []Wishlist{
			{
				Slug:        "diego-spicy",
				Title:       "Diego's Spicy Picks",
				Description: "Heat and tamarind.",
				IsPublic:    true,
				Items: []Item{
					{Name: "Takis Fuego", Type: "Chips/Crackers", Brand: "Barcel", Notes: ""},
					{Name: "Pelon Pelo Rico", Type: "Candy", Brand: "De La Rosa", Notes: "Tamarind"},
					{Name: "Jarritos", Type: "Beverages", Brand: "Jarritos", Notes: "Tamarind or mandarin"},
				},
			},
		},
	},
}

func Run(ctx context.Context, pool *pgxpool.Pool, searchClient *search.Client, password string) error {
	if password == "" {
		password = DefaultPassword
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	var createdUsers, createdWishlists, createdItems int

	for _, u := range Users {
		userID, created, err := ensureUser(ctx, pool, u, hash)
		if err != nil {
			return err
		}
		if created {
			createdUsers++
			log.Printf("created user %s (%s)", u.Username, u.Email)
		} else {
			log.Printf("user %s already exists, skipping", u.Username)
		}

		for _, w := range u.Wishlists {
			wishlistID, created, err := ensureWishlist(ctx, pool, userID, w)
			if err != nil {
				return err
			}
			if created {
				createdWishlists++
				log.Printf("  created wishlist %s", w.Slug)
			}

			n, err := ensureItems(ctx, pool, searchClient, wishlistID, w)
			if err != nil {
				return err
			}
			createdItems += n
		}
	}

	log.Printf(
		"seed complete: %d users, %d wishlists, %d items created (password: %s)",
		createdUsers, createdWishlists, createdItems, password,
	)
	return nil
}

func Remove(ctx context.Context, pool *pgxpool.Pool, searchClient *search.Client) error {
	var removedUsers int

	for _, u := range Users {
		var userID uuid.UUID
		err := pool.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, u.Email).Scan(&userID)
		if err != nil {
			log.Printf("user %s not found, skipping", u.Email)
			continue
		}

		if err := unindexUserWishlists(ctx, pool, searchClient, userID); err != nil {
			return fmt.Errorf("unindex wishlists for %s: %w", u.Email, err)
		}

		tag, err := pool.Exec(ctx, `DELETE FROM users WHERE id = $1 AND email = $2`, userID, u.Email)
		if err != nil {
			return fmt.Errorf("delete user %s: %w", u.Email, err)
		}
		if tag.RowsAffected() == 0 {
			log.Printf("user %s not deleted, skipping", u.Email)
			continue
		}

		removedUsers++
		log.Printf("removed user %s (%s)", u.Username, u.Email)
	}

	log.Printf("seed removal complete: %d users removed", removedUsers)
	return nil
}

func unindexUserWishlists(ctx context.Context, pool *pgxpool.Pool, searchClient *search.Client, userID uuid.UUID) error {
	if searchClient == nil {
		return nil
	}

	rows, err := pool.Query(ctx, `SELECT id FROM wishlists WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var wishlistID uuid.UUID
		if err := rows.Scan(&wishlistID); err != nil {
			return err
		}
		if err := searchClient.DeleteWishlistItems(ctx, wishlistID.String()); err != nil {
			return err
		}
	}
	return rows.Err()
}

func ensureUser(ctx context.Context, pool *pgxpool.Pool, u User, passwordHash string) (uuid.UUID, bool, error) {
	var userID uuid.UUID
	err := pool.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, u.Email).Scan(&userID)
	if err == nil {
		if err := ensureUserAvatar(ctx, pool, userID, u.AvatarURL); err != nil {
			return uuid.Nil, false, err
		}
		return userID, false, nil
	}

	var avatarURL *string
	if u.AvatarURL != "" {
		avatarURL = &u.AvatarURL
	}

	err = pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, display_name, country, username, email_verified, bio, avatar_url)
		VALUES ($1, $2, $3, $4, $5, TRUE, $6, $7)
		RETURNING id
	`, u.Email, passwordHash, u.DisplayName, u.Country, u.Username, u.Bio, avatarURL).Scan(&userID)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("create user %s: %w", u.Email, err)
	}
	return userID, true, nil
}

func ensureUserAvatar(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, avatarURL string) error {
	if avatarURL == "" {
		return nil
	}
	_, err := pool.Exec(ctx, `
		UPDATE users
		SET avatar_url = $2, avatar_key = NULL
		WHERE id = $1
		  AND (avatar_url IS NULL OR avatar_url = '')
		  AND avatar_key IS NULL
	`, userID, avatarURL)
	return err
}

func ensureWishlist(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, w Wishlist) (uuid.UUID, bool, error) {
	var wishlistID uuid.UUID
	err := pool.QueryRow(ctx, `SELECT id FROM wishlists WHERE slug = $1`, w.Slug).Scan(&wishlistID)
	if err == nil {
		return wishlistID, false, nil
	}

	err = pool.QueryRow(ctx, `
		INSERT INTO wishlists (user_id, slug, title, description, is_public)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, userID, w.Slug, w.Title, w.Description, w.IsPublic).Scan(&wishlistID)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("create wishlist %s: %w", w.Slug, err)
	}
	return wishlistID, true, nil
}

func ensureItems(ctx context.Context, pool *pgxpool.Pool, searchClient *search.Client, wishlistID uuid.UUID, w Wishlist) (int, error) {
	var existing int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM wishlist_items WHERE wishlist_id = $1`, wishlistID).Scan(&existing); err != nil {
		return 0, err
	}
	if existing > 0 {
		return 0, nil
	}

	created := 0
	for _, item := range w.Items {
		var itemID uuid.UUID
		err := pool.QueryRow(ctx, `
			INSERT INTO wishlist_items (wishlist_id, name, type, brand, notes)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`, wishlistID, item.Name, item.Type, item.Brand, item.Notes).Scan(&itemID)
		if err != nil {
			return created, fmt.Errorf("create item %q on %s: %w", item.Name, w.Slug, err)
		}
		created++
		indexItem(ctx, pool, searchClient, itemID)
	}
	return created, nil
}

func indexItem(ctx context.Context, pool *pgxpool.Pool, searchClient *search.Client, itemID uuid.UUID) {
	if searchClient == nil {
		return
	}

	var (
		wishlistID                   uuid.UUID
		name, itemType, brand, notes string
		wishlistSlug, wishlistTitle  string
		isPublic                     bool
		userID                       uuid.UUID
		username, displayName        string
		emailVerified                bool
		deactivatedAt                interface{}
	)
	err := pool.QueryRow(ctx, `
		SELECT wi.wishlist_id, wi.name, wi.type, wi.brand, wi.notes,
		       w.slug, w.title, w.is_public,
		       u.id, u.username, u.display_name, u.email_verified, u.deactivated_at
		FROM wishlist_items wi
		JOIN wishlists w ON w.id = wi.wishlist_id
		JOIN users u ON u.id = w.user_id
		WHERE wi.id = $1
	`, itemID).Scan(
		&wishlistID, &name, &itemType, &brand, &notes,
		&wishlistSlug, &wishlistTitle, &isPublic,
		&userID, &username, &displayName, &emailVerified, &deactivatedAt,
	)
	if err != nil || !isPublic || !emailVerified || deactivatedAt != nil {
		return
	}

	_ = searchClient.IndexItem(ctx, search.IndexDocument{
		ID:            itemID.String(),
		WishlistID:    wishlistID.String(),
		WishlistSlug:  wishlistSlug,
		WishlistTitle: wishlistTitle,
		UserID:        userID.String(),
		Username:      username,
		UserName:      displayName,
		Name:          name,
		Type:          itemType,
		Brand:         brand,
		Notes:         notes,
	})
}
