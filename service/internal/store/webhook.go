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

// WebhookDeliveryRow records one outbound webhook POST attempt.
type WebhookDeliveryRow struct {
	ID              int
	WebhookTargetID int
	Event           string
	URL             string
	Success         bool
	HTTPStatus      int
	ErrorMessage    string
	FiredAt         string
}
