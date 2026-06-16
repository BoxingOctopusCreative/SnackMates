package turnstile

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultSiteVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

var siteVerifyURL = defaultSiteVerifyURL

// SiteVerifyURLForTest overrides the siteverify endpoint during tests.
func SiteVerifyURLForTest(url string) string {
	prev := siteVerifyURL
	siteVerifyURL = url
	return prev
}

type siteVerifyResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

// Verify checks a Turnstile token with Cloudflare. When secret is empty, verification is skipped.
func Verify(ctx context.Context, secret, token, remoteIP string) error {
	if strings.TrimSpace(secret) == "" {
		return nil
	}
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("missing turnstile token")
	}

	data := url.Values{}
	data.Set("secret", secret)
	data.Set("response", token)
	if strings.TrimSpace(remoteIP) != "" {
		data.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, siteVerifyURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("create turnstile request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("turnstile request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read turnstile response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("turnstile verify returned status %d", resp.StatusCode)
	}

	var result siteVerifyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parse turnstile response: %w", err)
	}
	if !result.Success {
		return fmt.Errorf("turnstile verification failed: %v", result.ErrorCodes)
	}
	return nil
}
