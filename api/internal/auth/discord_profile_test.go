package auth

import "testing"

func TestDiscordDisplayName(t *testing.T) {
	global := "Snack Fan"
	tests := []struct {
		user discordUser
		want string
	}{
		{discordUser{Username: "snackfan", GlobalName: &global}, "Snack Fan"},
		{discordUser{Username: "snackfan"}, "snackfan"},
	}
	for _, tt := range tests {
		if got := discordDisplayName(tt.user); got != tt.want {
			t.Fatalf("discordDisplayName() = %q, want %q", got, tt.want)
		}
	}
}

func TestDiscordBio(t *testing.T) {
	bio := "  loves chips  "
	if got := discordBio(discordUser{Bio: &bio}); got != "loves chips" {
		t.Fatalf("discordBio() = %q", got)
	}
	if got := discordBio(discordUser{}); got != "" {
		t.Fatalf("discordBio() = %q, want empty", got)
	}
}

func TestProfileFromDiscordUserIncludesBio(t *testing.T) {
	bio := "snack enthusiast"
	profile := ProfileFromDiscordUser(discordUser{
		ID:       "1",
		Username: "snackfan",
		Bio:      &bio,
	})
	if profile.Bio != "snack enthusiast" {
		t.Fatalf("Bio = %q", profile.Bio)
	}
}

func TestDiscordAvatarURL(t *testing.T) {
	hash := "abc123"
	got := discordAvatarURL("123456789012345678", &hash)
	want := "https://cdn.discordapp.com/avatars/123456789012345678/abc123.png?size=256"
	if got != want {
		t.Fatalf("discordAvatarURL() = %q, want %q", got, want)
	}
}
