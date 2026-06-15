package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/yherda-opensource/yherda-cmd/internal/config"
)

const defaultAPIURL = "https://public.a.yherda.com"

type Client struct {
	workspace string
	creds     *config.Credentials
	http      *http.Client
}

func New(workspace string, creds *config.Credentials) *Client {
	httpClient := &http.Client{}
	if os.Getenv("YHERDA_API_URL") != "" {
		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			},
		}
	}
	return &Client{
		workspace: workspace,
		creds:     creds,
		http:      httpClient,
	}
}

func (c *Client) baseURL() string {
	base := os.Getenv("YHERDA_API_URL")
	if base == "" {
		base = defaultAPIURL
	}
	base = strings.TrimRight(base, "/")

	// If a custom URL is set, use it directly (local dev, staging, etc.)
	if os.Getenv("YHERDA_API_URL") != "" {
		return base + "/api"
	}

	// Default: route to tenant subdomain
	if c.workspace != "" {
		return fmt.Sprintf("https://%s.a.yherda.com/api", c.workspace)
	}
	return base + "/api"
}

func (c *Client) do(method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL()+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.creds != nil {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.creds.AccessToken))
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) Get(path string, out any) error {
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error %d: %s", resp.StatusCode, path)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) Post(path string, body any, out any) error {
	resp, err := c.do("POST", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(data))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
