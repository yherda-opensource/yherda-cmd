package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

func TestBaseURL_ExplicitAPIServer(t *testing.T) {
	creds := &config.Credentials{AccessToken: "tok"}

	c := New("https://community.yherda.test:8000", creds)
	if got, want := c.baseURL(), "https://community.yherda.test:8000/api"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBaseURL_ProductionAPIServer(t *testing.T) {
	creds := &config.Credentials{AccessToken: "tok"}

	c := New("https://community.a.yherda.com", creds)
	if got, want := c.baseURL(), "https://community.a.yherda.com/api"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBaseURL_PublicHost_Default(t *testing.T) {
	creds := &config.Credentials{AccessToken: "tok"}

	c := NewPublic(creds)
	if got, want := c.baseURL(), "https://public.a.yherda.com"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBaseURL_PublicHost_Override(t *testing.T) {
	t.Setenv("YHERDA_PUBLIC_HOST", "https://public.yherda.test:8000")
	creds := &config.Credentials{AccessToken: "tok"}

	c := NewPublic(creds)
	if got, want := c.baseURL(), "https://public.yherda.test:8000"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDo_401_TriggersRefreshAndRetry(t *testing.T) {
	var callCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := callCount.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer new-access" {
			t.Errorf("retry used wrong token: %q", auth)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{}`)
	}))
	defer srv.Close()

	creds := &config.Credentials{AccessToken: "old-access"}
	refreshCalled := 0
	c := New(srv.URL, creds)
	c.RefreshFunc = func() (*config.Credentials, error) {
		refreshCalled++
		return &config.Credentials{AccessToken: "new-access"}, nil
	}

	var out map[string]any
	if err := c.Get("/test", &out); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if callCount.Load() != 2 {
		t.Errorf("expected 2 HTTP calls, got %d", callCount.Load())
	}
	if refreshCalled != 1 {
		t.Errorf("expected RefreshFunc called once, got %d", refreshCalled)
	}
}

func TestDo_401_RefreshFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	creds := &config.Credentials{AccessToken: "old-access"}
	c := New(srv.URL, creds)
	c.RefreshFunc = func() (*config.Credentials, error) {
		return nil, fmt.Errorf("invalid_grant")
	}

	var out map[string]any
	err := c.Get("/test", &out)
	if err == nil {
		t.Fatal("expected error when refresh fails, got nil")
	}
}

func TestDo_401_NoRefreshFunc(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	creds := &config.Credentials{AccessToken: "old-access"}
	c := New(srv.URL, creds)
	// RefreshFunc intentionally not set.

	var out map[string]any
	err := c.Get("/test", &out)
	if err == nil {
		t.Fatal("expected 401 error with no RefreshFunc, got nil")
	}
}
