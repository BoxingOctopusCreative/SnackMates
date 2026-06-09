package handlers

import (
	"context"
	"net/url"
	"strings"

	"github.com/boxingoctopus/snackmates/api/internal/auth"
	"github.com/boxingoctopus/snackmates/api/internal/models"
	"github.com/boxingoctopus/snackmates/api/internal/storage"
)

func userProfileResponse(ctx context.Context, s *storage.Client, u *auth.UserRecord) map[string]interface{} {
	model := u.ToModel().(map[string]interface{})
	if s == nil {
		return model
	}
	avatarURL, err := s.ResolveObjectURL(ctx, u.AvatarURL, u.AvatarKey)
	if err == nil && avatarURL != "" {
		model["avatar_url"] = avatarURL
	}
	bannerURL, err := s.ResolveObjectURL(ctx, u.BannerURL, u.BannerKey)
	if err == nil && bannerURL != "" {
		model["banner_url"] = bannerURL
	}
	return model
}

func resolvePublicBannerURL(ctx context.Context, s *storage.Client, storedURL, key *string) *string {
	if s == nil {
		if storedURL != nil && strings.TrimSpace(*storedURL) != "" {
			return storedURL
		}
		return nil
	}
	bannerURL, err := s.ResolveObjectURL(ctx, storedURL, key)
	if err != nil || bannerURL == "" {
		return nil
	}
	return &bannerURL
}

func isAllowedBannerURL(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme != "https" {
		return false
	}
	host := strings.ToLower(parsed.Host)
	return host == "images.unsplash.com" || host == "plus.unsplash.com"
}

func enrichMatches(ctx context.Context, s *storage.Client, matches []models.SnackMatch) {
	if s == nil {
		return
	}
	for i := range matches {
		if matches[i].Mate == nil {
			continue
		}
		avatarURL, err := s.ResolveObjectURL(ctx, matches[i].Mate.AvatarURL, matches[i].Mate.AvatarKey)
		if err == nil && avatarURL != "" {
			matches[i].Mate.AvatarURL = &avatarURL
		}
	}
}
