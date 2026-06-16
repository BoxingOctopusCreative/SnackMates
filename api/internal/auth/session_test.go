package auth_test

import (
	"testing"
	"time"

	"github.com/boxingoctopus/snackmates/api/internal/auth"
)

func TestSessionTTL(t *testing.T) {
	if auth.SessionTTL(false) != 24*time.Hour {
		t.Fatalf("short session = %v", auth.SessionTTL(false))
	}
	if auth.SessionTTL(true) != 30*24*time.Hour {
		t.Fatalf("remembered session = %v", auth.SessionTTL(true))
	}
}

func TestSessionRemembered(t *testing.T) {
	if !auth.SessionRemembered(time.Now().Add(30*24*time.Hour)) {
		t.Fatal("expected remembered session")
	}
	if auth.SessionRemembered(time.Now().Add(12*time.Hour)) {
		t.Fatal("expected short-lived session")
	}
}

func TestCookieMaxAge(t *testing.T) {
	if auth.CookieMaxAge(time.Now().Add(30*24*time.Hour)) <= 0 {
		t.Fatal("expected persistent cookie max age")
	}
	if auth.CookieMaxAge(time.Now().Add(12*time.Hour)) != 0 {
		t.Fatal("expected session cookie max age")
	}
}
