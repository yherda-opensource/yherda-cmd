package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

// setupTest wires a temp home directory with credentials and config, and points
// YHERDA_API_URL at the provided test server so mustClient() works in tests.
func setupTest(t *testing.T, srv *httptest.Server) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("YHERDA_API_URL", srv.URL)

	if err := config.SaveCredentials(&config.Credentials{
		AccessToken: "test-token",
		TokenType:   "Bearer",
	}); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}
	if err := config.SaveConfig(&config.Config{ActiveWorkspace: "test-workspace"}); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
}

func jsonHandler(t *testing.T, expectPath string, payload any) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, expectPath) {
			t.Errorf("path: got %q, want suffix %q", r.URL.Path, expectPath)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload) //nolint:errcheck
	}
}

func runCmd(t *testing.T, args ...string) {
	t.Helper()
	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("command %v: %v", args, err)
	}
}

func TestIdeasList(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(t, "/storylines/", []map[string]any{{"id": 1, "name": "Test Idea"}}))
	defer srv.Close()
	setupTest(t, srv)
	runCmd(t, "ideas", "list")
}

func TestIdeasShow(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(t, "/storylines/42/", map[string]any{"id": 42, "name": "Test Idea"}))
	defer srv.Close()
	setupTest(t, srv)
	runCmd(t, "ideas", "show", "42")
}

func TestIdentitiesList(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(t, "/ideas/42/identities/", []map[string]any{{"id": 1, "name": "Alice"}}))
	defer srv.Close()
	setupTest(t, srv)
	runCmd(t, "identities", "list", "--idea", "42")
}

func TestIdentitiesShow(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(t, "/identities/7/", map[string]any{"id": 7, "name": "Alice"}))
	defer srv.Close()
	setupTest(t, srv)
	runCmd(t, "identities", "show", "7")
}

func TestArcsList(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(t, "/storylines/42/arcs/", []map[string]any{{"id": 3, "name": "Arc One"}}))
	defer srv.Close()
	setupTest(t, srv)
	runCmd(t, "arcs", "list", "--idea", "42")
}

func TestArcsShow(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(t, "/arcs/3/", map[string]any{"id": 3, "name": "Arc One"}))
	defer srv.Close()
	setupTest(t, srv)
	runCmd(t, "arcs", "show", "3")
}

func TestBeatsList(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(t, "/arcs/3/beats/", []map[string]any{{"id": 5, "name": "Beat One"}}))
	defer srv.Close()
	setupTest(t, srv)
	runCmd(t, "beats", "list", "--arc", "3")
}

func TestBeatsShow(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(t, "/beats/5/", map[string]any{"id": 5, "name": "Beat One"}))
	defer srv.Close()
	setupTest(t, srv)
	runCmd(t, "beats", "show", "5")
}

func TestIdeasList_Pretty(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(t, "/storylines/", []map[string]any{{"id": 1}}))
	defer srv.Close()
	setupTest(t, srv)

	// Capture stdout to verify indented output.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	runCmd(t, "--pretty", "ideas", "list")

	w.Close()
	os.Stdout = old

	var buf strings.Builder
	b := make([]byte, 4096)
	n, _ := r.Read(b)
	buf.Write(b[:n])

	out := buf.String()
	if !strings.Contains(out, "\n") || !strings.Contains(out, "  ") {
		t.Errorf("--pretty output does not look indented: %q", out)
	}
}
