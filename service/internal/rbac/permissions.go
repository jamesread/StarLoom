package rbac

const (
	PermissionAppAccess          = "app.access"
	PermissionUsersView          = "users.view"
	PermissionUsersCreate        = "users.create"
	PermissionUsersDelete        = "users.delete"
	PermissionUsersResetPassword = "users.reset-password"
	PermissionUserGroupsView     = "usergroups.view"
	PermissionUserGroupsManage   = "usergroups.manage"
	PermissionRbacView           = "rbac.view"
	PermissionRbacManage         = "rbac.manage"
	PermissionSystemSettings     = "system.settings"
	PermissionSystemLogs         = "system.logs"
	PermissionSystemImpersonate  = "system.impersonate"

	PermissionFamilyView         = "family.view"
	PermissionFamilyManage       = "family.manage"
	PermissionMembersManage      = "members.manage"
	PermissionMembersAvatar      = "members.avatar"
	PermissionStarsViewFamily    = "stars.view_family"
	PermissionStarsViewOwn       = "stars.view_own"
	PermissionStarsAward         = "stars.award"
	PermissionStarsRevoke        = "stars.revoke"
	PermissionRewardsManage      = "rewards.manage"
	PermissionRewardsView        = "rewards.view"
	PermissionRedemptionsApprove = "redemptions.approve"
	PermissionRedemptionsRequest = "redemptions.request"

	PermissionChoresManage    = "chores.manage"
	PermissionChoresComplete  = "chores.complete"
	PermissionChoresViewFamily = "chores.view_family"
)

const RoleSuperuser = "superuser"
const RoleMember = "member"
const RoleParent = "parent"
const RoleChild = "child"

const GroupEveryone = "Everyone"
const GroupAdministrators = "Administrators"
const GroupParents = "Parents"
const GroupChildren = "Children"
