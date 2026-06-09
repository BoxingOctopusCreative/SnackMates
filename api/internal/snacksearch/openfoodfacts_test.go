package snacksearch

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOFFClientUsesSearchALicious(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if got := r.URL.Query().Get("q"); got != "coffee crisp" {
			t.Fatalf("q = %q", got)
		}
		_ = json.NewEncoder(w).Encode(offSearchALiciousResponse{
			Hits: []json.RawMessage{
				json.RawMessage(`{"code":"123","product_name":"Coffee Crisp","brands":["Nestlé"],"categories":"Chocolate bars"}`),
			},
		})
	}))
	defer server.Close()

	client := newOFFClient(offClientOptions{SearchURL: server.URL})
	products, err := client.Search(context.Background(), "coffee crisp", 10)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if gotPath != "/search" {
		t.Fatalf("path = %q", gotPath)
	}
	if len(products) != 1 || products[0].Brands != "Nestlé" {
		t.Fatalf("unexpected products: %+v", products)
	}
}

func TestOFFClientFallsBackToLegacySearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/search":
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
		case "/cgi/search.pl":
			_ = json.NewEncoder(w).Encode(offSearchResponse{
				Products: []offProduct{{
					Code:        "456",
					ProductName: "Pocky",
					Brands:      "Glico",
					Categories:  "Biscuits and cakes",
				}},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newOFFClient(offClientOptions{
		SearchURL:     server.URL,
		LegacyBaseURL: server.URL,
	})
	products, err := client.Search(context.Background(), "pocky", 10)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
		if len(products) != 1 || products[0].ProductName != "Pocky" {
		t.Fatalf("unexpected products: %+v", products)
	}
}

func TestParseFlexibleString(t *testing.T) {
	if got := parseFlexibleString(json.RawMessage(`"Nestlé"`)); got != "Nestlé" {
		t.Fatalf("string = %q", got)
	}
	if got := parseFlexibleString(json.RawMessage(`["Nestlé","Mars"]`)); got != "Nestlé, Mars" {
		t.Fatalf("array = %q", got)
	}
}
