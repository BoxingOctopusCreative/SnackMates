package auth

import (
	"fmt"
	"strconv"
	"strings"
)

type DiscordProfile struct {
	DisplayName string
	Bio         string
	AvatarURL   string
	AvatarHash  *string
}

type discordUser struct {
	ID         string  `json:"id"`
	Username   string  `json:"username"`
	GlobalName *string `json:"global_name"`
	Avatar     *string `json:"avatar"`
	Bio        *string `json:"bio"`
	Email      string  `json:"email"`
}

func ProfileFromDiscordUser(du discordUser) DiscordProfile {
	hash := du.Avatar
	return DiscordProfile{
		DisplayName: discordDisplayName(du),
		Bio:         discordBio(du),
		AvatarURL:   discordAvatarURL(du.ID, hash),
		AvatarHash:  hash,
	}
}

func discordDisplayName(du discordUser) string {
	if du.GlobalName != nil {
		if name := strings.TrimSpace(*du.GlobalName); name != "" {
			return name
		}
	}
	return strings.TrimSpace(du.Username)
}

func discordBio(du discordUser) string {
	if du.Bio == nil {
		return ""
	}
	return strings.TrimSpace(*du.Bio)
}

func discordAvatarURL(userID string, avatar *string) string {
	if avatar != nil && *avatar != "" {
		ext := "png"
		if strings.HasPrefix(*avatar, "a_") {
			ext = "gif"
		}
		return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.%s?size=256", userID, *avatar, ext)
	}

	id, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return "https://cdn.discordapp.com/embed/avatars/0.png"
	}
	return fmt.Sprintf("https://cdn.discordapp.com/embed/avatars/%d.png", (id>>22)%6)
}
