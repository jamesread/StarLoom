package apprise

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultTimeout = 15 * time.Second
	maxRetries     = 3
	baseRetryDelay = 500 * time.Millisecond

	// PersonTagPrefix is the Apprise tag prefix; X is the family member (person) id.
	PersonTagPrefix = "starloom_uid_"
)

// Payload is the JSON body accepted by Apprise API /notify/ endpoints.
type Payload struct {
	Title string `json:"title,omitempty"`
	Body  string `json:"body"`
	Type  string `json:"type,omitempty"`
	Tag   string `json:"tag,omitempty"`
}

// PersonTag returns the Apprise tag for a family member (person) id.
func PersonTag(personID int) string {
	return fmt.Sprintf("%s%d", PersonTagPrefix, personID)
}

// JoinPersonTags builds a comma-separated OR tag expression for the given person ids.
func JoinPersonTags(personIDs []int) string {
	if len(personIDs) == 0 {
		return ""
	}
	parts := make([]string, 0, len(personIDs))
	for _, id := range personIDs {
		parts = append(parts, PersonTag(id))
	}
	return strings.Join(parts, ",")
}

// Notify POSTs payload to the Apprise API URL with retries. Empty URL is a no-op.
func Notify(client *http.Client, url string, payload Payload) error {
	if strings.TrimSpace(url) == "" {
		return nil
	}
	if client == nil {
		client = &http.Client{Timeout: defaultTimeout}
	}
	if payload.Type == "" {
		payload.Type = "info"
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal apprise payload: %w", err)
	}
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(baseRetryDelay * (1 << (attempt - 1)))
		}
		if err := postOnce(client, url, body); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	return fmt.Errorf("apprise notify failed after %d attempts: %w", maxRetries, lastErr)
}

func postOnce(client *http.Client, url string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}
