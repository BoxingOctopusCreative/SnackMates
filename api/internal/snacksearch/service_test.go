package snacksearch

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockOFFSearchHandler(t *testing.T, expectedQuery string, products []offProduct) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("q"); got != expectedQuery {
			t.Fatalf("q = %q, want %q", got, expectedQuery)
		}
		hits := make([]json.RawMessage, 0, len(products))
		for _, product := range products {
			raw, err := json.Marshal(product)
			if err != nil {
				t.Fatal(err)
			}
			hits = append(hits, raw)
		}
		_ = json.NewEncoder(w).Encode(offSearchALiciousResponse{Hits: hits})
	}
}

func TestServiceSearchWithoutAI(t *testing.T) {
	offServer := httptest.NewServer(mockOFFSearchHandler(t, "coffee crisp", []offProduct{{
		Code:        "123",
		ProductName: "Coffee Crisp",
		Brands:      "Nestlé",
		Categories:  "Chocolate stuffed wafers",
		ImageURL:    "https://example.com/coffee-crisp.jpg",
		Quantity:    "50 g",
	}}))
	defer offServer.Close()

	svc := New(Options{OFFSearchURL: offServer.URL})
	resp, err := svc.Search(context.Background(), "coffee crisp", 10)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if resp.AIAssisted {
		t.Fatal("expected ai_assisted=false without API key")
	}
	if len(resp.Results) != 1 {
		t.Fatalf("results len = %d", len(resp.Results))
	}
	if resp.Results[0].Type != "Candy" {
		t.Fatalf("type = %q", resp.Results[0].Type)
	}
}

func TestServiceSearchWithAI(t *testing.T) {
	var anthropicCalls int
	aiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		anthropicCalls++
		var text string
		switch anthropicCalls {
		case 1:
			text = `{"search_terms":"nestle coffee crisp chocolate bar","preferred_type":"Candy"}`
		case 2:
			text = `[{"code":"123","type":"Candy","relevance":0.95}]`
		default:
			t.Fatalf("unexpected anthropic call count: %d", anthropicCalls)
		}

		if r.URL.Path == "/v1/messages" {
			_ = json.NewEncoder(w).Encode(anthropicMessagesResponse{
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{{Type: "text", Text: text}},
			})
			return
		}
		t.Fatalf("unexpected path: %s", r.URL.Path)
	}))
	defer aiServer.Close()

	offServer := httptest.NewServer(mockOFFSearchHandler(t, "nestle coffee crisp chocolate bar", []offProduct{{
		Code:        "123",
		ProductName: "Coffee Crisp",
		Brands:      "Nestlé",
		Categories:  "Chocolate stuffed wafers",
	}}))
	defer offServer.Close()

	client := aiServer.Client()

	svc := New(Options{
		AnthropicAPIKey:  "test-key",
		AnthropicBaseURL: aiServer.URL,
		OFFSearchURL:     offServer.URL,
		HTTPClient:       client,
	})

	resp, err := svc.Search(context.Background(), "coffee crisp", 10)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if !resp.AIAssisted {
		t.Fatal("expected ai_assisted=true")
	}
	if resp.SearchTerms != "nestle coffee crisp chocolate bar" {
		t.Fatalf("search_terms = %q", resp.SearchTerms)
	}
	if len(resp.Results) != 1 || resp.Results[0].Relevance != 0.95 {
		t.Fatalf("unexpected results: %+v", resp.Results)
	}
	if anthropicCalls != 2 {
		t.Fatalf("anthropic calls = %d", anthropicCalls)
	}
}

func TestServiceSearchFallsBackToOVHWhenAnthropicFails(t *testing.T) {
	var ovhCalls int
	aiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/messages":
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		case "/chat/completions":
			ovhCalls++
			var text string
			if ovhCalls == 1 {
				text = `{"search_terms":"nestle coffee crisp","preferred_type":"Candy"}`
			} else {
				text = `[{"code":"123","type":"Candy","relevance":0.9}]`
			}
			_ = json.NewEncoder(w).Encode(openAIChatResponse{
				Choices: []struct {
					Message openAIMessage `json:"message"`
				}{{Message: openAIMessage{Role: "assistant", Content: text}}},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer aiServer.Close()

	offServer := httptest.NewServer(mockOFFSearchHandler(t, "nestle coffee crisp", []offProduct{{
		Code:        "123",
		ProductName: "Coffee Crisp",
		Brands:      "Nestlé",
		Categories:  "Chocolate stuffed wafers",
	}}))
	defer offServer.Close()

	svc := New(Options{
		AnthropicAPIKey:  "test-key",
		AnthropicBaseURL: aiServer.URL,
		OVHAIToken:       "ovh-token",
		OVHAIBaseURL:     aiServer.URL,
		OFFSearchURL:     offServer.URL,
		HTTPClient:       aiServer.Client(),
	})

	resp, err := svc.Search(context.Background(), "coffee crisp", 10)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if !resp.AIAssisted {
		t.Fatal("expected ai_assisted=true")
	}
	if resp.SearchTerms != "nestle coffee crisp" {
		t.Fatalf("search_terms = %q", resp.SearchTerms)
	}
	if ovhCalls != 2 {
		t.Fatalf("ovh calls = %d", ovhCalls)
	}
}

func TestServiceSearchWithOVHOnly(t *testing.T) {
	var ovhCalls int
	aiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		ovhCalls++
		text := `{"search_terms":"pocky sticks","preferred_type":"Candy"}`
		if ovhCalls == 2 {
			text = `[{"code":"456","type":"Candy","relevance":0.88}]`
		}
		_ = json.NewEncoder(w).Encode(openAIChatResponse{
			Choices: []struct {
				Message openAIMessage `json:"message"`
			}{{Message: openAIMessage{Role: "assistant", Content: text}}},
		})
	}))
	defer aiServer.Close()

	offServer := httptest.NewServer(mockOFFSearchHandler(t, "pocky sticks", []offProduct{{
		Code:        "456",
		ProductName: "Pocky",
		Brands:      "Glico",
		Categories:  "Biscuits and cakes",
	}}))
	defer offServer.Close()

	svc := New(Options{
		OVHAIToken:   "ovh-token",
		OVHAIBaseURL: aiServer.URL,
		OFFSearchURL: offServer.URL,
		HTTPClient:   aiServer.Client(),
	})

	resp, err := svc.Search(context.Background(), "pocky", 10)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if !resp.AIAssisted {
		t.Fatal("expected ai_assisted=true")
	}
	if resp.SearchTerms != "pocky sticks" {
		t.Fatalf("search_terms = %q", resp.SearchTerms)
	}
	if ovhCalls != 2 {
		t.Fatalf("ovh calls = %d", ovhCalls)
	}
}

func TestInferSnackType(t *testing.T) {
	tests := map[string]string{
		"Potato chips":   "Chips/Crackers",
		"Chocolate bars": "Candy",
		"Soft drinks":    "Beverages",
		"Cookies":        "Baked Goods",
		"Canned soup":    "Pantry",
	}
	for categories, want := range tests {
		if got := inferSnackType(categories); got != want {
			t.Fatalf("%q => %q, want %q", categories, got, want)
		}
	}
}
