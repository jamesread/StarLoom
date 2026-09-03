package store

const (
	NotificationTypeTest                = "test"
	NotificationTypeChoreCompleted      = "chore_completed"
	NotificationTypeRedemptionRequested = "redemption_requested"
)

// NotificationDeliveryRow records one outbound Apprise notification attempt.
type NotificationDeliveryRow struct {
	ID                   int
	FamilyID             int
	RecipientMemberID    int
	RecipientDisplayName string
	NotificationType     string
	Title                string
	Success              bool
	ErrorMessage         string
	SentAt               string
}
