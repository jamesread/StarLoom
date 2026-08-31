package store

import (
	"context"

	"github.com/jamesread/starapp/service/internal/rbac"
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

	// IAM — user accounts
	CountUserAccounts(ctx context.Context) (int, error)
	GetUserByUsername(ctx context.Context, username string) (*UserAccountRow, error)
	GetUserByID(ctx context.Context, id int) (*UserAccountRow, error)
	ListUserAccounts(ctx context.Context) ([]UserAccountRow, error)
	CreateUserAccount(ctx context.Context, username, passwordHash, createdBy string) (int, error)
	DeleteUserAccount(ctx context.Context, id int) error
	UpdateUserPassword(ctx context.Context, id int, passwordHash string) error

	// Sessions
	CreateSession(ctx context.Context, sid string, userID int, impersonatorID *int) error
	GetSessionBySID(ctx context.Context, sid string) (*SessionRow, error)
	DeleteSession(ctx context.Context, sid string) error
	DeleteSessionsForUser(ctx context.Context, userID int) error

	// API keys
	ListAPIKeysForUser(ctx context.Context, userID int) ([]APIKeyRow, error)
	CreateAPIKey(ctx context.Context, userID int, name, keyValue string, readOnly bool) (int, error)
	DeleteAPIKey(ctx context.Context, id, userID int) error
	GetUserByAPIKey(ctx context.Context, keyValue string) (*UserAccountRow, bool, error)
	TouchAPIKeyUsed(ctx context.Context, keyValue string) error

	// RBAC
	LoadEffectiveRBAC(ctx context.Context, userID int) (*rbac.EffectiveRBAC, error)
	EnsureRBACBootstrap(ctx context.Context) error
	EnsureUserInEveryoneGroup(ctx context.Context, userID int) error
	CountUsersWithSuperuserViaGroups(ctx context.Context) (int, error)

	ListRBACPermissions(ctx context.Context) ([]RBACPermissionRow, error)
	ListRBACRoles(ctx context.Context) ([]RBACRoleRow, error)
	GetRBACRole(ctx context.Context, id int) (*RBACRoleRow, error)
	CreateRBACRole(ctx context.Context, name, description string, permissionIDs []int) (int, error)
	UpdateRBACRole(ctx context.Context, id int, name, description string, permissionIDs []int) error
	DeleteRBACRole(ctx context.Context, id int) error
	SetRBACRolePermissions(ctx context.Context, roleID int, permissionIDs []int) error
	ListRolePermissionIDs(ctx context.Context, roleID int) ([]int, error)
	ListPermissionRoleNames(ctx context.Context, permissionID int) ([]string, error)
	GetUserRbacRoleNames(ctx context.Context, userID int) ([]string, error)
	GetUserGroupRbacRoleIDs(ctx context.Context, groupID int) ([]int, error)
	SetUserGroupRbacRoles(ctx context.Context, groupID int, roleIDs []int) error
	ListRbacRoleGroupNames(ctx context.Context, roleID int) ([]string, error)
	ListRbacRoleUsernames(ctx context.Context, roleID int) ([]string, error)
	GetMyPermissionsAudit(ctx context.Context, userID int) ([]string, []string, bool, []MyPermissionAuditRow, error)

	// User groups
	ListUserGroups(ctx context.Context) ([]UserGroupRow, error)
	GetUserGroupByName(ctx context.Context, name string) (*UserGroupRow, error)
	GetUserGroupByID(ctx context.Context, id int) (*UserGroupRow, error)
	CreateUserGroup(ctx context.Context, name string) (int, error)
	DeleteUserGroup(ctx context.Context, id int) error
	ListUserGroupMemberIDs(ctx context.Context, groupID int) ([]int, error)
	ListUserGroupIDsForUser(ctx context.Context, userID int) ([]int, error)
	SetUserGroupMembers(ctx context.Context, groupID int, userIDs []int) error

	GetUserPreferences(ctx context.Context, userID int) (*UserPreferencesRow, error)
	SaveUserPreferences(ctx context.Context, userID int, language string, sidebarEnabled bool) error

	EnsureUserInGroup(ctx context.Context, userID int, groupName string) error

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
}
