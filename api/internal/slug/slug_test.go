package slug

import "testing"

func TestUsernameFromName(t *testing.T) {
	if got := UsernameFromName("Boxing Octopus"); got != "boxingoctopus" {
		t.Fatalf("got %q", got)
	}
	if got := UsernameFromName("  "); got != "user" {
		t.Fatalf("got %q", got)
	}
}

func TestWishlistFromTitle(t *testing.T) {
	if got := WishlistFromTitle("Sweet treats"); got != "sweet-treats" {
		t.Fatalf("got %q", got)
	}
	if got := WishlistFromTitle("  "); got != "wishlist" {
		t.Fatalf("got %q", got)
	}
}

func TestUnique(t *testing.T) {
	taken := map[string]bool{"sweet-treats": true, "sweet-treats-2": true}
	got := Unique("sweet-treats", func(s string) bool { return taken[s] })
	if got != "sweet-treats-3" {
		t.Fatalf("got %q", got)
	}
}
