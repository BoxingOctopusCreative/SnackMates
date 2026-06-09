package email_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/boxingoctopus/snackmates/api/internal/config"
	"github.com/boxingoctopus/snackmates/api/internal/email"
)

func TestSendMailgunSuccess(t *testing.T) {
	var gotAuth, gotContentType, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/mg.example.com/messages" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "api" || pass != "test-key" {
			t.Fatalf("basic auth = %q/%q ok=%v", user, pass, ok)
		}
		gotAuth = user
		gotContentType = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"<20260307123456.1.ABCD@mg.example.com>","message":"Queued. Thank you."}`))
	}))
	defer srv.Close()

	cfg := config.Config{
		EmailProvider:     "mailgun",
		EmailFrom:         "SnackMates <noreply@mg.example.com>",
		MailgunAPIKey:     "test-key",
		MailgunDomain:     "mg.example.com",
		MailgunAPIBaseURL: srv.URL,
	}

	if err := email.Send(cfg, "user@example.com", "Hello", "<p>Hi</p>"); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if gotAuth != "api" {
		t.Fatalf("auth user = %q", gotAuth)
	}
	if gotContentType != "application/x-www-form-urlencoded" {
		t.Fatalf("content type = %q", gotContentType)
	}
	if !strings.Contains(gotBody, "from=SnackMates") || !strings.Contains(gotBody, "user%40example.com") {
		t.Fatalf("body = %q", gotBody)
	}
}

func TestSendMailgunMissingCredentials(t *testing.T) {
	cfg := config.Config{
		EmailProvider: "mailgun",
		EmailFrom:     "noreply@example.com",
	}
	if err := email.Send(cfg, "user@example.com", "Hello", "<p>Hi</p>"); err == nil {
		t.Fatal("expected error for missing mailgun credentials")
	}
}

func TestSendMailgunAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Forbidden"}`))
	}))
	defer srv.Close()

	cfg := config.Config{
		EmailProvider:     "mailgun",
		EmailFrom:         "noreply@example.com",
		MailgunAPIKey:     "bad-key",
		MailgunDomain:     "mg.example.com",
		MailgunAPIBaseURL: srv.URL,
	}
	if err := email.Send(cfg, "user@example.com", "Hello", "<p>Hi</p>"); err == nil {
		t.Fatal("expected mailgun API error")
	} else if !strings.Contains(err.Error(), "Forbidden") {
		t.Fatalf("err = %v", err)
	}
}
