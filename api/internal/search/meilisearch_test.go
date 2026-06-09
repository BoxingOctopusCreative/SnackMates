package search

import (
	"context"
	"testing"
)

func TestDisabledClientSkipsWritesAndSearchFails(t *testing.T) {
	ctx := context.Background()
	if err := Disabled.IndexItem(ctx, IndexDocument{ID: "1", Name: "test"}); err != nil {
		t.Fatalf("IndexItem: %v", err)
	}
	if err := Disabled.DeleteItem(ctx, "1"); err != nil {
		t.Fatalf("DeleteItem: %v", err)
	}
	hits, err := Disabled.SearchWishlistItems(ctx, "test", 10)
	if err != nil {
		t.Fatalf("SearchWishlistItems: %v", err)
	}
	if len(hits) != 0 {
		t.Fatalf("expected no hits from disabled client, got %d", len(hits))
	}
}
