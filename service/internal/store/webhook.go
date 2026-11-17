package store

// WebhookTargetRow is a webhook target with subscribed events (secret for dispatch only).
type WebhookTargetRow struct {
	URL     string
	Secret  string
	Created string
	Updated string
	Events  []string
	ID      int
	Enabled bool
}

// WebhookDeliveryRow records one outbound webhook POST attempt.
type WebhookDeliveryRow struct {
	Event           string
	URL             string
	ErrorMessage    string
	FiredAt         string
	ID              int
	WebhookTargetID int
	HTTPStatus      int
	Success         bool
}
