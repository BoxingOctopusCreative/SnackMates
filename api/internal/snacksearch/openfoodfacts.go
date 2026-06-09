package snacksearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultOFFSearchURL      = "https://search.openfoodfacts.org"
	defaultOFFLegacyBaseURL  = "https://ca.openfoodfacts.org"
	defaultOFFUserAgent      = "SnackMates/1.0 (notifications@snackmates.food)"
	offSearchFields          = "code,product_name,brands,categories,image_url,image_front_url,quantity"
)

type offClient struct {
	searchURL      string
	legacyBaseURL  string
	userAgent      string
	httpClient     *http.Client
}

type offSearchResponse struct {
	Products []offProduct `json:"products"`
}

type offSearchALiciousResponse struct {
	Hits []json.RawMessage `json:"hits"`
}

type offProduct struct {
	Code          string `json:"code"`
	ProductName   string `json:"product_name"`
	Brands        string `json:"brands"`
	Categories    string `json:"categories"`
	ImageURL      string `json:"image_url"`
	ImageFrontURL string `json:"image_front_url"`
	Quantity      string `json:"quantity"`
}

type offClientOptions struct {
	SearchURL     string
	LegacyBaseURL string
	UserAgent     string
	HTTPClient    *http.Client
}

func newOFFClient(opts offClientOptions) *offClient {
	searchURL := strings.TrimSpace(opts.SearchURL)
	if searchURL == "" {
		searchURL = defaultOFFSearchURL
	}
	legacyBaseURL := strings.TrimSpace(opts.LegacyBaseURL)
	if legacyBaseURL == "" {
		legacyBaseURL = defaultOFFLegacyBaseURL
	}
	userAgent := strings.TrimSpace(opts.UserAgent)
	if userAgent == "" {
		userAgent = defaultOFFUserAgent
	}
	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &offClient{
		searchURL:     strings.TrimRight(searchURL, "/"),
		legacyBaseURL: strings.TrimRight(legacyBaseURL, "/"),
		userAgent:     userAgent,
		httpClient:    httpClient,
	}
}

func (c *offClient) Search(ctx context.Context, terms string, limit int) ([]offProduct, error) {
	if limit <= 0 {
		limit = 20
	}

	products, err := c.searchALicious(ctx, terms, limit)
	if err == nil {
		return products, nil
	}

	legacyProducts, legacyErr := c.searchLegacy(ctx, terms, limit)
	if legacyErr == nil {
		return legacyProducts, nil
	}
	return nil, fmt.Errorf("openfoodfacts search failed: %w; legacy fallback: %v", err, legacyErr)
}

func (c *offClient) searchALicious(ctx context.Context, terms string, limit int) ([]offProduct, error) {
	query := url.Values{}
	query.Set("q", terms)
	query.Set("page_size", fmt.Sprintf("%d", limit))
	query.Set("fields", offSearchFields)
	query.Set("langs", "en,fr")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.searchURL+"/search?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search-a-licious request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search-a-licious status %d", resp.StatusCode)
	}

	var payload offSearchALiciousResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("search-a-licious decode: %w", err)
	}

	products := make([]offProduct, 0, len(payload.Hits))
	for _, raw := range payload.Hits {
		product, err := parseOFFProduct(raw)
		if err != nil {
			continue
		}
		if productDisplayName(product) == "" {
			continue
		}
		products = append(products, product)
	}
	return products, nil
}

func (c *offClient) searchLegacy(ctx context.Context, terms string, limit int) ([]offProduct, error) {
	endpoint := fmt.Sprintf("%s/cgi/search.pl", c.legacyBaseURL)
	query := url.Values{}
	query.Set("search_terms", terms)
	query.Set("json", "1")
	query.Set("page_size", fmt.Sprintf("%d", limit))
	query.Set("fields", offSearchFields)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("legacy request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("legacy status %d", resp.StatusCode)
	}

	var payload offSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("legacy decode: %w", err)
	}
	return payload.Products, nil
}

func (c *offClient) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
}

func parseOFFProduct(raw json.RawMessage) (offProduct, error) {
	var payload struct {
		Code          string          `json:"code"`
		ProductName   string          `json:"product_name"`
		Brands        json.RawMessage `json:"brands"`
		Categories    json.RawMessage `json:"categories"`
		ImageURL      string          `json:"image_url"`
		ImageFrontURL string          `json:"image_front_url"`
		Quantity      string          `json:"quantity"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return offProduct{}, err
	}
	return offProduct{
		Code:          payload.Code,
		ProductName:   payload.ProductName,
		Brands:        parseFlexibleString(payload.Brands),
		Categories:    parseFlexibleString(payload.Categories),
		ImageURL:      payload.ImageURL,
		ImageFrontURL: payload.ImageFrontURL,
		Quantity:      payload.Quantity,
	}, nil
}

func parseFlexibleString(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var value string
	if json.Unmarshal(raw, &value) == nil {
		return strings.TrimSpace(value)
	}
	var values []string
	if json.Unmarshal(raw, &values) == nil {
		return strings.TrimSpace(strings.Join(values, ", "))
	}
	return ""
}

func productImageURL(p offProduct) string {
	if p.ImageFrontURL != "" {
		return p.ImageFrontURL
	}
	return p.ImageURL
}

func productDisplayName(p offProduct) string {
	name := strings.TrimSpace(p.ProductName)
	if name != "" {
		return name
	}
	return strings.TrimSpace(p.Brands)
}
