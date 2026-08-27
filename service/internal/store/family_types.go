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
	ID        int
	Name      string
	CreatedAt string
}

type FamilyMemberRow struct {
	ID            int
	FamilyID      int
	UserAccountID *int
	DisplayName   string
	Role          string
	AvatarPath    string
	CreatedAt     string
	Username      string
}

type StarLedgerRow struct {
	ID                int
	FamilyID          int
	ChildMemberID     int
	Amount            int
	EntryType         string
	Note              string
	RelatedRewardID   *int
	CreatedByMemberID *int
	CreatedAt         string
}

type RewardRow struct {
	ID               int
	FamilyID         int
	Title            string
	Description      string
	CostStars        int
	Active           bool
	ApprovalRequired bool
}

type RedemptionRow struct {
	ID                 int
	FamilyID           int
	ChildMemberID      int
	RewardID           int
	StarsSpent         int
	Status             string
	LedgerEntryID      *int
	CreatedAt          string
	ResolvedAt         string
	ResolvedByMemberID *int
	FulfilledAt        string
	RewardTitle        string
	ChildDisplayName   string
}

type ChildHomeSummaryRow struct {
	Member    FamilyMemberRow
	Balance   int
	LastAward *StarLedgerRow
}
