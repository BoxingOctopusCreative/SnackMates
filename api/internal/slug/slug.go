package slug

import (
	"regexp"
	"strings"
)

var (
	nonUsernameChars = regexp.MustCompile(`[^a-z0-9]+`)
	nonWishlistChars = regexp.MustCompile(`[^a-z0-9-]+`)
	multiDash        = regexp.MustCompile(`-+`)
)

func UsernameFromName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = nonUsernameChars.ReplaceAllString(name, "")
	if name == "" {
		return "user"
	}
	return name
}

func WishlistFromTitle(title string) string {
	title = strings.ToLower(strings.TrimSpace(title))
	title = strings.ReplaceAll(title, " ", "-")
	title = nonWishlistChars.ReplaceAllString(title, "")
	title = multiDash.ReplaceAllString(title, "-")
	title = strings.Trim(title, "-")
	if title == "" {
		return "wishlist"
	}
	return title
}

func IsValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 32 {
		return false
	}
	for _, r := range username {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			continue
		}
		return false
	}
	return true
}

func Unique(base string, taken func(string) bool) string {
	if !taken(base) {
		return base
	}
	for i := 2; ; i++ {
		candidate := base + "-" + itoa(i)
		if !taken(candidate) {
			return candidate
		}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
