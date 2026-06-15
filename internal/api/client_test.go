package api

import (
	"testing"

	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

func TestBaseURL_Dev(t *testing.T) {
	t.Setenv("YHERDA_DOMAIN_ROOT", "yherda.test:8000")
	creds := &config.Credentials{AccessToken: "tok"}

	c := &Client{workspace: "myworkspace", creds: creds}
	if got, want := c.baseURL(), "https://myworkspace.yherda.test:8000/api"; got != want {
		t.Errorf("workspace: got %q, want %q", got, want)
	}

	c2 := &Client{creds: creds}
	if got, want := c2.baseURL(), "https://public.yherda.test:8000/api"; got != want {
		t.Errorf("public: got %q, want %q", got, want)
	}
}

func TestBaseURL_Staging(t *testing.T) {
	t.Setenv("YHERDA_DOMAIN_ROOT", "staging.a.yherda.com")
	creds := &config.Credentials{AccessToken: "tok"}

	c := &Client{workspace: "myworkspace", creds: creds}
	if got, want := c.baseURL(), "https://myworkspace.staging.a.yherda.com/api"; got != want {
		t.Errorf("workspace: got %q, want %q", got, want)
	}

	c2 := &Client{creds: creds}
	if got, want := c2.baseURL(), "https://public.staging.a.yherda.com/api"; got != want {
		t.Errorf("public: got %q, want %q", got, want)
	}
}

func TestBaseURL_Production(t *testing.T) {
	creds := &config.Credentials{AccessToken: "tok"}

	c := &Client{workspace: "myworkspace", creds: creds}
	if got, want := c.baseURL(), "https://myworkspace.a.yherda.com/api"; got != want {
		t.Errorf("workspace: got %q, want %q", got, want)
	}

	c2 := &Client{creds: creds}
	if got, want := c2.baseURL(), "https://public.a.yherda.com/api"; got != want {
		t.Errorf("public: got %q, want %q", got, want)
	}
}
