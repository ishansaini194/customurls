package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const serverURL = "https://customurls.in"

var httpClient = &http.Client{Timeout: 10 * time.Second}

// ── Shorten ──────────────────────────────────────────────────────────

type shortenRequest struct {
	URL   string `json:"url"`
	Alias string `json:"alias,omitempty"`
}

type ShortenResponse struct {
	Short  string `json:"short"`
	URL    string `json:"url"`
	Expiry int    `json:"expiry"`
	Error  string `json:"error,omitempty"`
}

// shorten posts to /shorten and returns the parsed response.
func shorten(url, alias string) (*ShortenResponse, error) {
	body, err := json.Marshal(shortenRequest{URL: url, Alias: alias})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := httpClient.Post(serverURL+"/shorten", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var data ShortenResponse
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(raw))
	}

	if resp.StatusCode != http.StatusOK {
		if data.Error != "" {
			return nil, fmt.Errorf("%s", data.Error)
		}
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return &data, nil
}

// ── Stats ────────────────────────────────────────────────────────────

type StatsResponse struct {
	ShortID     string     `json:"short_id"`
	Hits        int        `json:"hits"`
	OriginalURL string     `json:"original_url"`
	CreatedAt   *time.Time `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	Error       string     `json:"error,omitempty"`
}

// stats fetches /stats/:shortID and returns the parsed response.
func stats(shortID string) (*StatsResponse, error) {
	resp, err := httpClient.Get(serverURL + "/stats/" + shortID)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var data StatsResponse
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(raw))
	}

	if resp.StatusCode != http.StatusOK {
		if data.Error != "" {
			return nil, fmt.Errorf("%s", data.Error)
		}
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("short URL not found")
		}
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return &data, nil
}
