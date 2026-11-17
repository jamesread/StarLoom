package store

type ChoreNotificationSubscriptionRow struct {
	ChildMemberID      *int
	ChoreID            *int
	CreatedAt          string
	ID                 int
	FamilyID           int
	SubscriberMemberID int
}
