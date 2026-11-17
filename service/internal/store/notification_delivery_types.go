package store

const (
	NotificationTypeTest                = "test"
	NotificationTypeChoreCompleted      = "chore_completed"
	NotificationTypeRedemptionRequested = "redemption_requested"
)

// NotificationDeliveryRow records one outbound Apprise notification attempt.
type NotificationDeliveryRow struct {
	RecipientDisplayName string
	NotificationType     string
	Title                string
	ErrorMessage         string
	SentAt               string
	ID                   int
	FamilyID             int
	RecipientMemberID    int
	Success              bool
}
