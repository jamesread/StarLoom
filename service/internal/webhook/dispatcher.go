package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jamesread/starapp/service/internal/store"
)

var SupportedEvents = []string{
	"stars.awarded",
	"redemption.requested",
	"redemption.resolved",
}

func NormalizeEvent(event string) (string, error) {
	event = strings.TrimSpace(event)
	for _, e := range SupportedEvents {
		if event == e {
			return event, nil
		}
	}
	return "", fmt.Errorf("unsupported webhook event")
}

func NormalizeEvents(events []string) ([]string, error) {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(events))
	for _, raw := range events {
		e, err := NormalizeEvent(raw)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[e]; ok {
			continue
		}
		seen[e] = struct{}{}
		out = append(out, e)
	}
	return out, nil
}

func NormalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("webhook URL must be non-empty")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("webhook URL is not valid")
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", fmt.Errorf("webhook URL scheme must be http or https")
	}
	if strings.TrimSpace(u.Host) == "" {
		return "", fmt.Errorf("webhook URL host is required")
	}
	return raw, nil
}

func Signature(payloadJSON, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payloadJSON))
	return hex.EncodeToString(mac.Sum(nil))
}

type Dispatcher struct {
	Store  store.Store
	Client *http.Client
}

func (d *Dispatcher) Dispatch(ctx context.Context, event string, data map[string]any) {
	event, err := NormalizeEvent(event)
	if err != nil {
		return
	}
	payload := map[string]any{
		"event":     event,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	for k, v := range data {
		payload[k] = v
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	targets, err := d.Store.EnabledTargetsForEvent(ctx, event)
	if err != nil {
		return
	}
	client := d.httpClient()
	for _, wh := range targets {
		d.postWebhook(ctx, client, event, wh, body)
	}
}

func (d *Dispatcher) httpClient() *http.Client {
	if d.Client != nil {
		return d.Client
	}
	return &http.Client{Timeout: 2 * time.Second}
}

func (d *Dispatcher) postWebhook(ctx context.Context, client *http.Client, event string, wh store.WebhookTargetRow, body []byte) {
	if _, err := NormalizeURL(wh.URL); err != nil {
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, wh.URL, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-StarApp-Event", event)
	req.Header.Set("X-StarApp-Signature", "sha256="+Signature(string(body), wh.Secret))
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	_ = resp.Body.Close()
}
