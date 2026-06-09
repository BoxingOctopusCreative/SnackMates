package snacksearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultAnthropicModel = "claude-haiku-4-5-20251001"
	defaultOVHModel       = "Mistral-Nemo-Instruct-2407"
	defaultOVHBaseURL     = "https://oai.endpoints.kepler.ai.cloud.ovh.net/v1"
)

type queryPlan struct {
	SearchTerms   string  `json:"search_terms"`
	PreferredType *string `json:"preferred_type"`
}

type classifiedProduct struct {
	Code      string  `json:"code"`
	Type      string  `json:"type"`
	Relevance float64 `json:"relevance"`
}

type llmProvider interface {
	enabled() bool
	name() string
	complete(ctx context.Context, prompt string) (string, error)
}

type assistant struct {
	providers []llmProvider
}

func newAssistant(providers ...llmProvider) *assistant {
	enabled := make([]llmProvider, 0, len(providers))
	for _, p := range providers {
		if p.enabled() {
			enabled = append(enabled, p)
		}
	}
	return &assistant{providers: enabled}
}

func (a *assistant) enabled() bool {
	return len(a.providers) > 0
}

func (a *assistant) complete(ctx context.Context, prompt string) (string, error) {
	if !a.enabled() {
		return "", fmt.Errorf("no llm providers configured")
	}

	var errs []string
	for _, provider := range a.providers {
		text, err := provider.complete(ctx, prompt)
		if err == nil {
			return text, nil
		}
		errs = append(errs, fmt.Sprintf("%s: %v", provider.name(), err))
	}
	return "", fmt.Errorf("all llm providers failed: %s", strings.Join(errs, "; "))
}

func (a *assistant) planQuery(ctx context.Context, query string, snackTypes []string) (queryPlan, error) {
	prompt := fmt.Sprintf(`You help users find packaged snack products on OpenFoodFacts.

Snack types (use exactly one of these labels when assigning a type):
%s

User query: %q

Return ONLY valid JSON with this shape:
{"search_terms":"optimized OpenFoodFacts search phrase","preferred_type":"SnackType or null"}

Rules:
- search_terms should be concise product-focused keywords (brand, product name, snack type).
- preferred_type is null unless the user clearly wants one snack category.
- Do not include markdown or commentary.`, strings.Join(snackTypes, ", "), query)

	text, err := a.complete(ctx, prompt)
	if err != nil {
		return queryPlan{}, err
	}

	var plan queryPlan
	if err := decodeJSON(text, &plan); err != nil {
		return queryPlan{}, err
	}
	if strings.TrimSpace(plan.SearchTerms) == "" {
		plan.SearchTerms = query
	}
	if plan.PreferredType != nil {
		normalized := normalizeSnackType(*plan.PreferredType, snackTypes)
		plan.PreferredType = &normalized
	}
	return plan, nil
}

func (a *assistant) classifyProducts(
	ctx context.Context,
	query string,
	snackTypes []string,
	products []offProduct,
) ([]classifiedProduct, error) {
	if len(products) == 0 {
		return nil, nil
	}

	type compactProduct struct {
		Code       string `json:"code"`
		Name       string `json:"name"`
		Brand      string `json:"brand"`
		Categories string `json:"categories"`
	}

	compact := make([]compactProduct, 0, len(products))
	for _, p := range products {
		compact = append(compact, compactProduct{
			Code:       p.Code,
			Name:       productDisplayName(p),
			Brand:      strings.TrimSpace(p.Brands),
			Categories: strings.TrimSpace(p.Categories),
		})
	}

	payload, err := json.Marshal(compact)
	if err != nil {
		return nil, err
	}

	prompt := fmt.Sprintf(`You rank OpenFoodFacts snack products for a user search.

Snack types (assign exactly one per product):
%s

User query: %q
Products JSON:
%s

Return ONLY valid JSON array:
[{"code":"barcode","type":"SnackType","relevance":0.0}]

Rules:
- relevance is 0-1 (1 = perfect match).
- Omit products with relevance below 0.2.
- Sort by relevance descending.
- type must be one of the snack types above.
- Do not include markdown or commentary.`,
		strings.Join(snackTypes, ", "), query, string(payload))

	text, err := a.complete(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var classified []classifiedProduct
	if err := decodeJSON(text, &classified); err != nil {
		return nil, err
	}
	return classified, nil
}

func decodeJSON[T any](text string, out *T) error {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(text, "```")
		text = strings.TrimSpace(text)
	}
	if err := json.Unmarshal([]byte(text), out); err != nil {
		return fmt.Errorf("parse model JSON: %w", err)
	}
	return nil
}

type anthropicProvider struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

type anthropicMessagesRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicMessagesResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func newAnthropicProvider(apiKey, model, baseURL string, httpClient *http.Client) *anthropicProvider {
	if strings.TrimSpace(model) == "" {
		model = defaultAnthropicModel
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://api.anthropic.com"
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &anthropicProvider{
		apiKey:     strings.TrimSpace(apiKey),
		model:      model,
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

func (p *anthropicProvider) enabled() bool {
	return p.apiKey != ""
}

func (p *anthropicProvider) name() string {
	return "anthropic"
}

func (p *anthropicProvider) complete(ctx context.Context, prompt string) (string, error) {
	body, err := json.Marshal(anthropicMessagesRequest{
		Model:     p.model,
		MaxTokens: 1024,
		Messages: []anthropicMessage{{
			Role:    "user",
			Content: prompt,
		}},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var payload anthropicMessagesResponse
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	if payload.Error != nil {
		return "", fmt.Errorf("api error: %s", payload.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}
	if len(payload.Content) == 0 || payload.Content[0].Text == "" {
		return "", fmt.Errorf("empty response")
	}
	return payload.Content[0].Text, nil
}

type openAIProvider struct {
	apiKey     string
	model      string
	baseURL    string
	provider   string
	httpClient *http.Client
}

type openAIChatRequest struct {
	Model       string          `json:"model"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature"`
	Messages    []openAIMessage `json:"messages"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func newOpenAIProvider(name, apiKey, model, baseURL string, httpClient *http.Client) *openAIProvider {
	if strings.TrimSpace(model) == "" {
		model = defaultOVHModel
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultOVHBaseURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &openAIProvider{
		apiKey:     strings.TrimSpace(apiKey),
		model:      model,
		baseURL:    strings.TrimRight(baseURL, "/"),
		provider:   name,
		httpClient: httpClient,
	}
}

func (p *openAIProvider) enabled() bool {
	return p.apiKey != ""
}

func (p *openAIProvider) name() string {
	return p.provider
}

func (p *openAIProvider) complete(ctx context.Context, prompt string) (string, error) {
	body, err := json.Marshal(openAIChatRequest{
		Model:       p.model,
		MaxTokens:   1024,
		Temperature: 0,
		Messages: []openAIMessage{{
			Role:    "user",
			Content: prompt,
		}},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var payload openAIChatResponse
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	if payload.Error != nil {
		return "", fmt.Errorf("api error: %s", payload.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}
	if len(payload.Choices) == 0 || payload.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("empty response")
	}
	return payload.Choices[0].Message.Content, nil
}
