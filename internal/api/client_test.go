package api

import (
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
