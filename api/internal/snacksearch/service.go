package snacksearch

import (
	"context"
	"net/http"
	"strings"

	"github.com/boxingoctopus/snackmates/api/internal/models"
)

type ProductHit struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	Brand      string  `json:"brand"`
	Type       string  `json:"type"`
	Categories string  `json:"categories,omitempty"`
	ImageURL   string  `json:"image_url,omitempty"`
	Quantity   string  `json:"quantity,omitempty"`
	Relevance  float64 `json:"relevance,omitempty"`
}

type Response struct {
	Query       string       `json:"query"`
	SearchTerms string       `json:"search_terms"`
	AIAssisted  bool         `json:"ai_assisted"`
	Results     []ProductHit `json:"results"`
}

type Service struct {
	off       *offClient
	assistant *assistant
}

type Options struct {
	AnthropicAPIKey  string
	AnthropicModel   string
	AnthropicBaseURL string
	OVHAIToken       string
	OVHAIModel       string
	OVHAIBaseURL     string
	OFFSearchURL     string
	OFFBaseURL       string
	HTTPClient       *http.Client
}

func New(opts Options) *Service {
	httpClient := opts.HTTPClient
	return &Service{
		off: newOFFClient(offClientOptions{
			SearchURL:     opts.OFFSearchURL,
			LegacyBaseURL: opts.OFFBaseURL,
			HTTPClient:    httpClient,
		}),
		assistant: newAssistant(
			newAnthropicProvider(opts.AnthropicAPIKey, opts.AnthropicModel, opts.AnthropicBaseURL, httpClient),
			newOpenAIProvider("ovh-mistral", opts.OVHAIToken, opts.OVHAIModel, opts.OVHAIBaseURL, httpClient),
		),
	}
}

func (s *Service) Search(ctx context.Context, query string, limit int) (Response, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return Response{Results: []ProductHit{}}, nil
	}
	if limit <= 0 {
		limit = 20
	}

	searchTerms := query
	aiAssisted := false

	if s.assistant.enabled() {
		plan, err := s.assistant.planQuery(ctx, query, models.ValidSnackTypes)
		if err == nil {
			searchTerms = plan.SearchTerms
			aiAssisted = true
		}
	}

	products, err := s.off.Search(ctx, searchTerms, limit*2)
	if err != nil {
		return Response{}, err
	}

	results := s.buildResults(ctx, query, products, aiAssisted, limit)
	return Response{
		Query:       query,
		SearchTerms: searchTerms,
		AIAssisted:  aiAssisted,
		Results:     results,
	}, nil
}

func (s *Service) buildResults(ctx context.Context, query string, products []offProduct, aiAssisted bool, limit int) []ProductHit {
	if len(products) == 0 {
		return []ProductHit{}
	}

	byCode := make(map[string]offProduct, len(products))
	for _, p := range products {
		if p.Code == "" {
			continue
		}
		byCode[p.Code] = p
	}

	if aiAssisted && s.assistant.enabled() {
		classified, err := s.assistant.classifyProducts(ctx, query, models.ValidSnackTypes, products)
		if err == nil && len(classified) > 0 {
			return hitsFromClassification(classified, byCode, limit)
		}
	}

	hits := make([]ProductHit, 0, len(products))
	for _, p := range products {
		name := productDisplayName(p)
		if name == "" {
			continue
		}
		hits = append(hits, ProductHit{
			Code:       p.Code,
			Name:       name,
			Brand:      strings.TrimSpace(p.Brands),
			Type:       inferSnackType(p.Categories),
			Categories: strings.TrimSpace(p.Categories),
			ImageURL:   productImageURL(p),
			Quantity:   strings.TrimSpace(p.Quantity),
			Relevance:  0.5,
		})
		if len(hits) >= limit {
			break
		}
	}
	return hits
}

func hitsFromClassification(classified []classifiedProduct, byCode map[string]offProduct, limit int) []ProductHit {
	hits := make([]ProductHit, 0, len(classified))
	for _, item := range classified {
		p, ok := byCode[item.Code]
		if !ok {
			continue
		}
		name := productDisplayName(p)
		if name == "" {
			continue
		}
		snackType := normalizeSnackType(item.Type, models.ValidSnackTypes)
		if snackType == "" {
			snackType = inferSnackType(p.Categories)
		}
		hits = append(hits, ProductHit{
			Code:       p.Code,
			Name:       name,
			Brand:      strings.TrimSpace(p.Brands),
			Type:       snackType,
			Categories: strings.TrimSpace(p.Categories),
			ImageURL:   productImageURL(p),
			Quantity:   strings.TrimSpace(p.Quantity),
			Relevance:  item.Relevance,
		})
		if len(hits) >= limit {
			break
		}
	}
	return hits
}

func normalizeSnackType(value string, snackTypes []string) string {
	value = strings.TrimSpace(value)
	for _, t := range snackTypes {
		if strings.EqualFold(value, t) {
			return t
		}
	}
	return ""
}

func inferSnackType(categories string) string {
	lower := strings.ToLower(categories)
	switch {
	case strings.Contains(lower, "chip"), strings.Contains(lower, "cracker"), strings.Contains(lower, "crisp"):
		return "Chips/Crackers"
	case strings.Contains(lower, "candy"), strings.Contains(lower, "chocolate"), strings.Contains(lower, "gumm"):
		return "Candy"
	case strings.Contains(lower, "cookie"), strings.Contains(lower, "cake"), strings.Contains(lower, "pastry"), strings.Contains(lower, "bread"), strings.Contains(lower, "biscuit"):
		return "Baked Goods"
	case strings.Contains(lower, "beverage"), strings.Contains(lower, "drink"), strings.Contains(lower, "soda"), strings.Contains(lower, "juice"):
		return "Beverages"
	default:
		return "Pantry"
	}
}
