package cache

import (
	"context"
	"testing"
	"time"
)

func TestDisabledClientReturnsUnavailable(t *testing.T) {
	ctx := context.Background()
	if err := Disabled.Set(ctx, "k", "v", time.Minute); err == nil {
		t.Fatal("expected error from disabled Set")
	}
	if _, err := Disabled.Get(ctx, "k"); err == nil {
		t.Fatal("expected error from disabled Get")
	}
	if err := Disabled.Delete(ctx, "k"); err == nil {
		t.Fatal("expected error from disabled Delete")
	}
	if err := Disabled.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}
