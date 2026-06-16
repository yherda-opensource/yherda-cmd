package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

const (
	clientID    = "yherda-cmd-public-client"
	defaultBase = "https://public.a.yherda.com"
	scope       = "read write"
)

func baseURL() string {
	if v := os.Getenv("YHERDA_PUBLIC_HOST"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return defaultBase
}

func httpClient() *http.Client {
	if os.Getenv("YHERDA_INSECURE") == "1" {
		return &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			},
		}
	}
	return http.DefaultClient
}

func Login() (*config.Credentials, error) {
	verifier, challenge, err := makePKCEPair()
	if err != nil {
		return nil, fmt.Errorf("generating PKCE pair: %w", err)
	}

	// Start local callback server on a random port.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("starting callback server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	srv := &http.Server{Handler: mux}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		authErr := r.URL.Query().Get("error")
		if authErr != "" {
			fmt.Fprintf(w, "<html><body><h2>Authentication failed: %s</h2><p>You can close this window.</p></body></html>", authErr)
			errCh <- fmt.Errorf("authorization error: %s", authErr)
			return
		}
		if code == "" {
			fmt.Fprintf(w, "<html><body><h2>No code received.</h2><p>You can close this window.</p></body></html>")
			errCh <- fmt.Errorf("no authorization code in callback")
			return
		}
		fmt.Fprintf(w, "<html><body><h2>Authentication complete.</h2><p>You can close this window.</p></body></html>")
		codeCh <- code
	})

	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	authURL := buildAuthURL(redirectURI, challenge)
	fmt.Printf("Opening browser to authenticate...\n%s\n\n", authURL)
	openBrowser(authURL)

	var code string
	select {
	case code = <-codeCh:
	case err = <-errCh:
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		srv.Shutdown(ctx) //nolint:errcheck
		return nil, err
	case <-time.After(5 * time.Minute):
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		srv.Shutdown(ctx) //nolint:errcheck
		return nil, fmt.Errorf("login timed out after 5 minutes")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	srv.Shutdown(ctx) //nolint:errcheck

	creds, err := exchangeCode(code, verifier, redirectURI)
	if err != nil {
		return nil, fmt.Errorf("exchanging code: %w", err)
	}
	return creds, nil
}

func buildAuthURL(redirectURI, challenge string) string {
	params := url.Values{
		"response_type":         {"code"},
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"scope":                 {scope},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
	}
	return baseURL() + "/oauth/authorize/?" + params.Encode()
}

func exchangeCode(code, verifier, redirectURI string) (*config.Credentials, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {clientID},
		"code_verifier": {verifier},
	}

	req, err := http.NewRequest(http.MethodPost, baseURL()+"/oauth/token/", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]any
		json.NewDecoder(resp.Body).Decode(&errBody) //nolint:errcheck
		return nil, fmt.Errorf("token endpoint returned %d: %v", resp.StatusCode, errBody)
	}

	var tok struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return nil, err
	}

	return &config.Credentials{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		TokenType:    tok.TokenType,
	}, nil
}

func RefreshTokens(refreshToken string) (*config.Credentials, error) {
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {clientID},
	}

	req, err := http.NewRequest(http.MethodPost, baseURL()+"/oauth/token/", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]any
		json.NewDecoder(resp.Body).Decode(&errBody) //nolint:errcheck
		return nil, fmt.Errorf("token refresh failed (%d): %v", resp.StatusCode, errBody)
	}

	var tok struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return nil, err
	}

	return &config.Credentials{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		TokenType:    tok.TokenType,
	}, nil
}

func makePKCEPair() (verifier, challenge string, err error) {
	buf := make([]byte, 48)
	if _, err = rand.Read(buf); err != nil {
		return
	}
	verifier = base64.RawURLEncoding.EncodeToString(buf)
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	return
}
