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

type shortenRequest struct {
	URL   string `json:"url"`
	Alias string `json:"alias,omitempty"`
}

type shortenResponse struct {
	Short  string `json:"short"`
	URL    string `json:"url"`
	Expiry int    `json:"expiry"`
	Error  string `json:"error,omitempty"`
}

// shorten posts to /shorten and returns the short URL or an error.
func shorten(url, alias string) (string, error) {
	body, err := json.Marshal(shortenRequest{URL: url, Alias: alias})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(serverURL+"/shorten", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var data shortenResponse
	if err := json.Unmarshal(raw, &data); err != nil {
		return "", fmt.Errorf("server returned %d: %s", resp.StatusCode, string(raw))
	}

	if resp.StatusCode != http.StatusOK {
		if data.Error != "" {
			return "", fmt.Errorf("%s", data.Error)
		}
		return "", fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return data.Short, nil
}
