package turnstile_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/boxingoctopus/snackmates/api/internal/turnstile"
)

func TestVerifySkipsWhenSecretEmpty(t *testing.T) {
	if err := turnstile.Verify(context.Background(), "", "token", "127.0.0.1"); err != nil {
		t.Fatalf("Verify: %v", err)
	}
}

func TestVerifyRequiresTokenWhenSecretSet(t *testing.T) {
	if err := turnstile.Verify(context.Background(), "secret", "", "127.0.0.1"); err == nil {
		t.Fatal("expected error for missing token")
	}
}

func TestVerifySuccess(t *testing.T) {
	var gotSecret, gotToken, gotRemoteIP string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		values, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatalf("parse body: %v", err)
		}
		gotSecret = values.Get("secret")
		gotToken = values.Get("response")
		gotRemoteIP = values.Get("remoteip")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer srv.Close()

	origURL := turnstile.SiteVerifyURLForTest(srv.URL)
	defer turnstile.SiteVerifyURLForTest(origURL)

	if err := turnstile.Verify(context.Background(), "test-secret", "test-token", "203.0.113.1"); err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if gotSecret != "test-secret" || gotToken != "test-token" || gotRemoteIP != "203.0.113.1" {
		t.Fatalf("request = secret:%q token:%q remoteip:%q", gotSecret, gotToken, gotRemoteIP)
	}
}

func TestVerifyFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":false,"error-codes":["invalid-input-response"]}`))
	}))
	defer srv.Close()

	origURL := turnstile.SiteVerifyURLForTest(srv.URL)
	defer turnstile.SiteVerifyURLForTest(origURL)

	if err := turnstile.Verify(context.Background(), "test-secret", "bad-token", ""); err == nil {
		t.Fatal("expected verification error")
	}
}
