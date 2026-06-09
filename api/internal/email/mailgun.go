package email

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/boxingoctopus/snackmates/api/internal/config"
)

const defaultMailgunAPIBaseURL = "https://api.mailgun.net"

func sendMailgun(cfg config.Config, to, subject, body string) error {
	apiKey := strings.TrimSpace(cfg.MailgunAPIKey)
	domain := strings.TrimSpace(cfg.MailgunDomain)
	if apiKey == "" || domain == "" {
		return fmt.Errorf("mailgun api key and domain are required when email provider is mailgun")
	}

	baseURL := strings.TrimRight(strings.TrimSpace(cfg.MailgunAPIBaseURL), "/")
	if baseURL == "" {
		baseURL = defaultMailgunAPIBaseURL
	}

	form := url.Values{}
	form.Set("from", cfg.EmailFrom)
	form.Set("to", to)
	form.Set("subject", subject)
	form.Set("html", body)

	endpoint := baseURL + "/v3/" + domain + "/messages"
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth("api", apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if msg := mailgunErrorMessage(respBody); msg != "" {
		return fmt.Errorf("mailgun send failed (%d): %s", resp.StatusCode, msg)
	}
	return fmt.Errorf("mailgun send failed with status %d", resp.StatusCode)
}

func mailgunErrorMessage(body []byte) string {
	var payload struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return strings.TrimSpace(string(body))
	}
	return strings.TrimSpace(payload.Message)
}
