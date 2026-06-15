package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

func testClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	t.Setenv("YHERDA_API_URL", srv.URL)
	creds := &config.Credentials{AccessToken: "test-token", TokenType: "Bearer"}
	return New("test-workspace", creds), srv
}

func TestGet_Success(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
			t.Errorf("Authorization header: got %q", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{{"id": 1, "name": "My Idea"}}) //nolint:errcheck
	})

	var result []map[string]any
	if err := client.Get("/storylines/", &result); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0]["name"] != "My Idea" {
		t.Errorf("name: got %v", result[0]["name"])
	}
}

func TestGet_404(t *testing.T) {
	client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	var result any
	err := client.Get("/storylines/999/", &result)
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

func TestGet_PathRouting(t *testing.T) {
	paths := []struct {
		path string
		desc string
	}{
		{"/storylines/", "ideas list"},
		{"/storylines/42/", "ideas show"},
		{"/ideas/42/identities/", "identities list"},
		{"/identities/7/", "identities show"},
		{"/storylines/42/arcs/", "arcs list"},
		{"/arcs/3/", "arcs show"},
		{"/arcs/3/beats/", "beats list"},
		{"/beats/5/", "beats show"},
	}

	for _, tc := range paths {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			var gotPath string
			client, _ := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode([]any{}) //nolint:errcheck
			})

			var result any
			if err := client.Get(tc.path, &result); err != nil {
				t.Fatalf("Get(%q): %v", tc.path, err)
			}
			if gotPath != "/api"+tc.path {
				t.Errorf("path: got %q, want %q", gotPath, "/api"+tc.path)
			}
		})
	}
}

func TestBaseURL_WorkspaceSubdomainSubstitution(t *testing.T) {
	t.Setenv("YHERDA_API_URL", "https://public.yherda.test:8000")
	creds := &config.Credentials{AccessToken: "tok"}

	client := New("myworkspace", creds)
	got := client.baseURL()
	want := "https://myworkspace.yherda.test:8000/api"
	if got != want {
		t.Errorf("baseURL: got %q, want %q", got, want)
	}
}

func TestBaseURL_WorkspaceSubdomainProduction(t *testing.T) {
	creds := &config.Credentials{AccessToken: "tok"}
	client := New("myworkspace", creds)
	got := client.baseURL()
	want := "https://myworkspace.a.yherda.com/api"
	if got != want {
		t.Errorf("baseURL: got %q, want %q", got, want)
	}
}

func TestGet_AuthHeader(t *testing.T) {
	creds := &config.Credentials{AccessToken: "my-secret-token", TokenType: "Bearer"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer my-secret-token" {
			t.Errorf("Authorization: got %q, want %q", got, "Bearer my-secret-token")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{}) //nolint:errcheck
	}))
	t.Cleanup(srv.Close)
	t.Setenv("YHERDA_API_URL", srv.URL)

	client := New("test-workspace", creds)
	var result any
	if err := client.Get("/storylines/1/", &result); err != nil {
		t.Fatalf("Get: %v", err)
	}
}
