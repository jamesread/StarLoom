package store

import (
	"context"

	iamstore "github.com/jamesread/armature-iam/store"
)

// Store abstracts persistence for StarApp domain data.
type Store interface {
	Close() error
	Ping(ctx context.Context) error
	HasMigration(ctx context.Context, id string) (bool, error)
	LatestMigration(ctx context.Context) (string, error)
	ListCvars(ctx context.Context) ([]CvarRow, error)
	FindCvar(ctx context.Context, key string) (*CvarRow, error)
	InsertCvarIfMissing(ctx context.Context, row CvarRow) error
	UpdateCvar(ctx context.Context, key string, valueInt int, valueString string) error
	ListWebhookTargets(ctx context.Context) ([]WebhookTargetRow, error)
	FindWebhookTarget(ctx context.Context, id int) (*WebhookTargetRow, error)
	EnabledTargetsForEvent(ctx context.Context, event string) ([]WebhookTargetRow, error)
	CreateWebhookTarget(ctx context.Context, url, secret string, events []string, enabled bool) (int, error)
	UpdateWebhookTarget(ctx context.Context, id int, url, secret string, events []string, enabled bool, clearSecret bool) error
	DeleteWebhookTarget(ctx context.Context, id int) error
	InsertWebhookDelivery(ctx context.Context, row WebhookDeliveryRow) (int, error)
	ListWebhookDeliveries(ctx context.Context, limit int) ([]WebhookDeliveryRow, error)

	iamstore.Store
	SeedDomainRBAC(ctx context.Context) error

	GetUserPreferences(ctx context.Context, userID int) (*UserPreferencesRow, error)
	SaveUserPreferences(ctx context.Context, userID int, language string, sidebarEnabled bool) error

	CountFamilies(ctx context.Context) (int, error)
	GetFamilyByID(ctx context.Context, id int) (*FamilyRow, error)
	GetFirstFamily(ctx context.Context) (*FamilyRow, error)
	CreateFamily(ctx context.Context, name string) (int, error)
	UpdateFamilyName(ctx context.Context, id int, name string) error

	GetMemberByID(ctx context.Context, id int) (*FamilyMemberRow, error)
	GetMemberByAccountID(ctx context.Context, accountID int) (*FamilyMemberRow, error)
	ListMembersByFamily(ctx context.Context, familyID int) ([]FamilyMemberRow, error)
	CreateMember(ctx context.Context, familyID int, displayName, role string, accountID *int, starColor string) (int, error)
	UpdateMember(ctx context.Context, id int, displayName, starColor string) error
	SetMemberUserAccount(ctx context.Context, id int, accountID int) error
	DeleteMember(ctx context.Context, id int) error
	SetMemberAvatarPath(ctx context.Context, id int, path string) error

	GetMemberBalance(ctx context.Context, memberID int) (int, error)
	ListLedger(ctx context.Context, memberID, limit int) ([]StarLedgerRow, error)
	GetLastAward(ctx context.Context, memberID int) (*StarLedgerRow, error)
	InsertLedgerEntry(ctx context.Context, row StarLedgerRow) (int, error)

	ListRewards(ctx context.Context, familyID int, includeInactive bool) ([]RewardRow, error)
	GetRewardByID(ctx context.Context, id int) (*RewardRow, error)
	CreateReward(ctx context.Context, familyID int, title, description string, costStars int, approvalRequired bool, availabilityExpression string) (int, error)
	UpdateReward(ctx context.Context, id int, title, description string, costStars int, active, approvalRequired bool, availabilityExpression string) error
	DeactivateReward(ctx context.Context, id int) error

	ListRedemptions(ctx context.Context, familyID int, status string) ([]RedemptionRow, error)
	GetRedemptionByID(ctx context.Context, id int) (*RedemptionRow, error)
	CreateRedemption(ctx context.Context, familyID, childMemberID, rewardID, starsSpent int, status string, ledgerEntryID *int) (int, error)
	ResolveRedemption(ctx context.Context, id int, status string, resolvedByMemberID int, ledgerEntryID *int) error
	CountPendingRedemptions(ctx context.Context, familyID int) (int, error)
	CountApprovedRedemptionsForMemberRewardBetween(ctx context.Context, childMemberID, rewardID int, startDate, endDate string) (int, error)

	ListChores(ctx context.Context, familyID int, starChartID int, includeInactive bool) ([]ChoreWithAssignments, error)
	GetChoreByID(ctx context.Context, id int) (*ChoreWithAssignments, error)
	CreateChore(ctx context.Context, familyID, starChartID int, title string, starReward, weekdayMask int, childMemberIDs []int) (int, error)
	UpdateChore(ctx context.Context, id int, starChartID int, title string, starReward, weekdayMask int, active bool, childMemberIDs []int) error
	DeactivateChore(ctx context.Context, id int) error

	ListStarCharts(ctx context.Context, familyID int, includeInactive bool) ([]StarChartRow, error)
	GetStarChartByID(ctx context.Context, id int) (*StarChartRow, error)
	GetDefaultStarChartID(ctx context.Context, familyID int) (int, error)
	CreateStarChart(ctx context.Context, familyID int, name string, sortOrder int) (int, error)
	UpdateStarChart(ctx context.Context, id int, name string, sortOrder int, active bool) error
	DeleteStarChart(ctx context.Context, id int) error
	CountChoresForStarChart(ctx context.Context, starChartID int) (int, error)
	CountChoresForStarChartAndMember(ctx context.Context, starChartID, memberID int) (int, error)

	ListChorePauses(ctx context.Context, familyID int) ([]ChorePauseRow, error)
	CreateChorePause(ctx context.Context, familyID int, startDate, endDate, reason string) (int, error)
	DeleteChorePause(ctx context.Context, id int) error
	IsDatePaused(ctx context.Context, familyID int, date string) (bool, error)

	GetAssignment(ctx context.Context, choreID, childMemberID int) (*ChoreAssignmentRow, error)
	GetCompletion(ctx context.Context, assignmentID int, date string) (*ChoreCompletionRow, error)
	InsertChoreCompletion(ctx context.Context, assignmentID int, date string, ledgerEntryID int) (int, error)
	DeleteChoreCompletion(ctx context.Context, id int) error
	ListCompletionsForWeek(ctx context.Context, familyID int, weekStart, weekEnd string) ([]WeeklyChartCompletion, error)
	ListChoreLedgerEntryIDs(ctx context.Context, familyID int) ([]int, error)
	ListBonusStarsForWeek(ctx context.Context, familyID int, weekStart, weekEnd string, choreLedgerIDs []int) (map[int]map[string]int, error)

	ListChoreNotificationSubscriptions(ctx context.Context, subscriberMemberID int) ([]ChoreNotificationSubscriptionRow, error)
	ReplaceChoreNotificationSubscriptions(ctx context.Context, familyID, subscriberMemberID int, subs []ChoreNotificationSubscriptionRow) error
	MatchingChoreNotificationSubscribers(ctx context.Context, familyID, childMemberID, choreID int) ([]int, error)

	InsertNotificationDelivery(ctx context.Context, row NotificationDeliveryRow) (int, error)
	ListNotificationDeliveries(ctx context.Context, limit int) ([]NotificationDeliveryRow, error)
}
