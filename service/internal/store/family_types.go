package store

const (
	MemberRoleParent = "parent"
	MemberRoleChild  = "child"

	LedgerTypeAward  = "award"
	LedgerTypeRevoke = "revoke"
	LedgerTypeRedeem = "redeem"

	RedemptionPending  = "pending"
	RedemptionApproved = "approved"
	RedemptionRejected = "rejected"
)

type FamilyRow struct {
	Name      string
	CreatedAt string
	ID        int
}

type FamilyMemberRow struct {
	UserAccountID *int
	DisplayName   string
	Role          string
	AvatarPath    string
	StarColor     string
	CreatedAt     string
	Username      string
	ID            int
	FamilyID      int
}

type StarLedgerRow struct {
	RelatedRewardID   *int
	CreatedByMemberID *int
	EntryType         string
	Note              string
	CreatedAt         string
	ID                int
	FamilyID          int
	ChildMemberID     int
	Amount            int
}

type RewardRow struct {
	Title                  string
	Description            string
	AvailabilityExpression string
	ID                     int
	FamilyID               int
	CostStars              int
	Active                 bool
	ApprovalRequired       bool
	MultiSpend             bool
}

type RedemptionRow struct {
	ResolvedByMemberID *int
	LedgerEntryID      *int
	Status             string
	CreatedAt          string
	ResolvedAt         string
	FulfilledAt        string
	RewardTitle        string
	ChildDisplayName   string
	FamilyID           int
	ChildMemberID      int
	RewardID           int
	StarsSpent         int
	ID                 int
}

type ChildHomeSummaryRow struct {
	LastAward *StarLedgerRow
	Member    FamilyMemberRow
	Balance   int
}
