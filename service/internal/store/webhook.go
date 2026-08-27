package store

// WebhookTargetRow is a webhook target with subscribed events (secret for dispatch only).
type WebhookTargetRow struct {
	ID      int
	URL     string
	Secret  string
	Enabled bool
	Created string
	Updated string
	Events  []string
}
