package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestMakePKCEPair(t *testing.T) {
	verifier, challenge, err := makePKCEPair()
	if err != nil {
		t.Fatalf("makePKCEPair: %v", err)
	}
	if len(verifier) == 0 {
		t.Error("verifier is empty")
	}
	if len(challenge) == 0 {
		t.Error("challenge is empty")
	}
	if verifier == challenge {
		t.Error("verifier and challenge must differ")
	}

	// Each call must produce a unique pair.
	v2, c2, err := makePKCEPair()
	if err != nil {
		t.Fatalf("makePKCEPair second call: %v", err)
	}
	if verifier == v2 {
		t.Error("expected unique verifiers across calls")
	}
	if challenge == c2 {
		t.Error("expected unique challenges across calls")
	}
}

func TestBuildAuthURL(t *testing.T) {
	t.Setenv("YHERDA_API_URL", "https://public.example.com")

	redirectURI := "http://127.0.0.1:9999/callback"
	challenge := "test-challenge"

	raw := buildAuthURL(redirectURI, challenge)
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("buildAuthURL produced invalid URL: %v", err)
	}

	q := parsed.Query()
	checks := map[string]string{
		"response_type":         "code",
		"client_id":             clientID,
		"redirect_uri":          redirectURI,
		"scope":                 scope,
		"code_challenge":        challenge,
		"code_challenge_method": "S256",
	}
	for key, want := range checks {
		if got := q.Get(key); got != want {
			t.Errorf("param %q: got %q, want %q", key, got, want)
		}
	}

	if !strings.HasPrefix(raw, "https://public.example.com") {
		t.Errorf("URL does not use YHERDA_API_URL base: %s", raw)
	}
}

func TestExchangeCode_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
			t.Errorf("unexpected Content-Type: %s", ct)
		}
		r.ParseForm() //nolint:errcheck
		if r.FormValue("grant_type") != "authorization_code" {
			t.Errorf("unexpected grant_type: %s", r.FormValue("grant_type"))
		}
		if r.FormValue("code") != "test-code" {
			t.Errorf("unexpected code: %s", r.FormValue("code"))
		}
		if r.FormValue("code_verifier") != "test-verifier" {
			t.Errorf("unexpected code_verifier: %s", r.FormValue("code_verifier"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{ //nolint:errcheck
			"access_token":  "access-abc",
			"refresh_token": "refresh-xyz",
			"token_type":    "Bearer",
		})
	}))
	defer srv.Close()

	t.Setenv("YHERDA_API_URL", srv.URL)

	creds, err := exchangeCode("test-code", "test-verifier", "http://127.0.0.1:9999/callback")
	if err != nil {
		t.Fatalf("exchangeCode: %v", err)
	}
	if creds.AccessToken != "access-abc" {
		t.Errorf("access_token: got %q, want %q", creds.AccessToken, "access-abc")
	}
	if creds.RefreshToken != "refresh-xyz" {
		t.Errorf("refresh_token: got %q, want %q", creds.RefreshToken, "refresh-xyz")
	}
	if creds.TokenType != "Bearer" {
		t.Errorf("token_type: got %q, want %q", creds.TokenType, "Bearer")
	}
}

func TestExchangeCode_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"}) //nolint:errcheck
	}))
	defer srv.Close()

	t.Setenv("YHERDA_API_URL", srv.URL)

	_, err := exchangeCode("bad-code", "bad-verifier", "http://127.0.0.1:9999/callback")
	if err == nil {
		t.Fatal("expected error from 400 response, got nil")
	}
}

func TestExchangeCode_MismatchedVerifier(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"}) //nolint:errcheck
	}))
	defer srv.Close()

	t.Setenv("YHERDA_API_URL", srv.URL)

	_, err := exchangeCode("test-code", "wrong-verifier", "http://127.0.0.1:9999/callback")
	if err == nil {
		t.Fatal("expected error for mismatched verifier, got nil")
	}
}
