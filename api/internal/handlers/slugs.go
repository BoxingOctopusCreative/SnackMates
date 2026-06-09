package handlers

import (
	"context"
	"strings"

	"github.com/boxingoctopus/snackmates/api/internal/slug"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func allocateWishlistSlug(ctx context.Context, pool *pgxpool.Pool, title string, excludeID *uuid.UUID) (string, error) {
	base := slug.WishlistFromTitle(title)
	return slug.Unique(base, func(candidate string) bool {
		var exists bool
		var err error
		if excludeID != nil {
			err = pool.QueryRow(ctx, `
				SELECT EXISTS(SELECT 1 FROM wishlists WHERE slug = $1 AND id <> $2)
			`, candidate, *excludeID).Scan(&exists)
		} else {
			err = pool.QueryRow(ctx, `
				SELECT EXISTS(SELECT 1 FROM wishlists WHERE slug = $1)
			`, candidate).Scan(&exists)
		}
		return err != nil || exists
	}), nil
}

func lookupWishlistID(ctx context.Context, pool *pgxpool.Pool, wishlistSlug string) (uuid.UUID, error) {
	var id uuid.UUID
	err := pool.QueryRow(ctx, `SELECT id FROM wishlists WHERE slug = $1`, strings.TrimSpace(wishlistSlug)).Scan(&id)
	return id, err
}
