package apprise

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jamesread/starapp/service/internal/buildinfo"
)

const (
	defaultTimeout = 15 * time.Second
	maxRetries     = 3
	baseRetryDelay = 500 * time.Millisecond

	// PersonTagPrefix is the Apprise tag prefix; X is the family member (person) id.
	PersonTagPrefix = "starloom_uid_"
)

// Payload is the JSON body accepted by Apprise API /notify/{key} endpoints.
type Payload struct {
	Title  string `json:"title,omitempty"`
	Body   string `json:"body"`
	Type   string `json:"type,omitempty"`
	Format string `json:"format,omitempty"`
	Tag    string `json:"tag,omitempty"`
}

type logResponse struct {
	Error   *string    `json:"error"`
	Details [][]string `json:"details"`
}

// UserAgent returns the HTTP User-Agent StarLoom sends to Apprise.
func UserAgent() string {
	version := strings.TrimSpace(buildinfo.Version)
	if version == "" {
		version = "dev"
	}
	return "StarLoom/" + version
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

// ValidateNotifyURL reports whether url targets a persistent Apprise configuration key.
func ValidateNotifyURL(raw string) error {
	key, ok := persistentNotifyKey(raw)
	if !ok {
		return fmt.Errorf("apprise URL must include a configuration key (e.g. http://apprise:8000/notify/mykey); bare /notify requires urls in each request, which StarLoom does not send")
	}
	if key == "" {
		return fmt.Errorf("apprise URL missing configuration key after /notify/")
	}
	return nil
}

func persistentNotifyKey(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", false
	}
	path := strings.TrimSuffix(u.Path, "/")
	if path == "" || path == "/notify" || strings.HasSuffix(path, "/notify") {
		return "", false
	}
	const marker = "/notify/"
	idx := strings.LastIndex(path, marker)
	if idx < 0 {
		return "", false
	}
	key := path[idx+len(marker):]
	if key == "" || strings.Contains(key, "/") {
		return "", false
	}
	return key, true
}

// Notify POSTs payload to the Apprise API URL with retries. Empty URL is a no-op.
func Notify(client *http.Client, notifyURL string, payload Payload) error {
	notifyURL = strings.TrimSpace(notifyURL)
	if notifyURL == "" {
		return nil
	}
	if err := ValidateNotifyURL(notifyURL); err != nil {
		return err
	}
	if client == nil {
		client = &http.Client{Timeout: defaultTimeout}
	}
	if payload.Type == "" {
		payload.Type = "info"
	}
	if payload.Format == "" {
		payload.Format = "text"
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
		if err := postOnce(client, notifyURL, body); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	return fmt.Errorf("apprise notify failed after %d attempts: %w", maxRetries, lastErr)
}

func postOnce(client *http.Client, notifyURL string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, notifyURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", UserAgent())

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if readErr != nil {
		return readErr
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var parsed logResponse
		if len(respBody) > 0 && json.Unmarshal(respBody, &parsed) == nil {
			if parsed.Error != nil && strings.TrimSpace(*parsed.Error) != "" {
				return fmt.Errorf("apprise: %s", strings.TrimSpace(*parsed.Error))
			}
		}
		return nil
	case http.StatusNoContent:
		return fmt.Errorf("apprise: no notification targets matched (HTTP 204); verify the configuration key exists in Apprise and tag filters match configured services")
	case http.StatusFailedDependency:
		return appriseErrorFromBody(respBody, "one or more notifications could not be sent")
	case http.StatusBadRequest:
		return appriseErrorFromBody(respBody, "invalid notification request")
	case http.StatusNotAcceptable:
		return fmt.Errorf("apprise: recursion limit reached (HTTP 406)")
	case http.StatusRequestHeaderFieldsTooLarge:
		return fmt.Errorf("apprise: request payload too large (HTTP 431)")
	default:
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		if msg := strings.TrimSpace(string(respBody)); msg != "" {
			return fmt.Errorf("apprise: HTTP %d: %s", resp.StatusCode, msg)
		}
		return fmt.Errorf("apprise: HTTP %d", resp.StatusCode)
	}
}

func appriseErrorFromBody(respBody []byte, fallback string) error {
	var parsed logResponse
	if len(respBody) > 0 && json.Unmarshal(respBody, &parsed) == nil {
		if parsed.Error != nil && strings.TrimSpace(*parsed.Error) != "" {
			return fmt.Errorf("apprise: %s", strings.TrimSpace(*parsed.Error))
		}
	}
	if msg := strings.TrimSpace(string(respBody)); msg != "" {
		return fmt.Errorf("apprise: %s", msg)
	}
	return fmt.Errorf("apprise: %s", fallback)
}
