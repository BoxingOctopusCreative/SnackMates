package handlers

import (
	"context"

	"github.com/boxingoctopus/snackmates/api/internal/matching"
	"github.com/boxingoctopus/snackmates/api/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func matchPkgGet(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) ([]models.SnackMatch, error) {
	return matching.GetMatchesForUser(ctx, pool, userID)
}

func matchPkgRun(ctx context.Context, pool *pgxpool.Pool) ([]models.SnackMatch, error) {
	return matching.PairUsers(ctx, pool)
}
