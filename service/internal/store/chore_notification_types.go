package store

type ChoreNotificationSubscriptionRow struct {
	ID                 int
	FamilyID           int
	SubscriberMemberID int
	ChildMemberID      *int
	ChoreID            *int
	CreatedAt          string
}
