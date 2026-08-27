package store

import (
	"context"
	"sort"
	"time"
)

type memoryWebhook struct {
	nextID int
	byID   map[int]WebhookTargetRow
}

func (m *Memory) webhookState() *memoryWebhook {
	if m.webhooks == nil {
		m.webhooks = &memoryWebhook{byID: map[int]WebhookTargetRow{}}
	}
	return m.webhooks
}

func (m *Memory) ListWebhookTargets(_ context.Context) ([]WebhookTargetRow, error) {
	st := m.webhookState()
	out := make([]WebhookTargetRow, 0, len(st.byID))
	for _, row := range st.byID {
		copy := row
		copy.Secret = ""
		out = append(out, copy)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (m *Memory) FindWebhookTarget(_ context.Context, id int) (*WebhookTargetRow, error) {
	row, ok := m.webhookState().byID[id]
	if !ok {
		return nil, nil
	}
	copy := row
	return &copy, nil
}

func (m *Memory) EnabledTargetsForEvent(_ context.Context, event string) ([]WebhookTargetRow, error) {
	out := make([]WebhookTargetRow, 0)
	for _, row := range m.webhookState().byID {
		if !row.Enabled {
			continue
		}
		for _, e := range row.Events {
			if e == event {
				out = append(out, row)
				break
			}
		}
	}
	return out, nil
}

func (m *Memory) CreateWebhookTarget(_ context.Context, url, secret string, events []string, enabled bool) (int, error) {
	st := m.webhookState()
	st.nextID++
	now := time.Now().UTC().Format(time.RFC3339)
	st.byID[st.nextID] = WebhookTargetRow{
		ID: st.nextID, URL: url, Secret: secret, Enabled: enabled,
		Created: now, Updated: now, Events: append([]string(nil), events...),
	}
	return st.nextID, nil
}

func (m *Memory) UpdateWebhookTarget(_ context.Context, id int, url, secret string, events []string, enabled bool, clearSecret bool) error {
	st := m.webhookState()
	cur, ok := st.byID[id]
	if !ok {
		return errCvarNotFound
	}
	if url != "" {
		cur.URL = url
	}
	if !clearSecret && secret != "" {
		cur.Secret = secret
	}
	cur.Enabled = enabled
	cur.Events = append([]string(nil), events...)
	cur.Updated = time.Now().UTC().Format(time.RFC3339)
	st.byID[id] = cur
	return nil
}

func (m *Memory) DeleteWebhookTarget(_ context.Context, id int) error {
	st := m.webhookState()
	if _, ok := st.byID[id]; !ok {
		return errCvarNotFound
	}
	delete(st.byID, id)
	return nil
}
