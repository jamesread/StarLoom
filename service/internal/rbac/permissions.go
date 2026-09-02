package rbac

import iamrbac "github.com/jamesread/armature-iam/rbac"

const (
	PermissionAppAccess          = iamrbac.PermissionAppAccess
	PermissionUsersView          = iamrbac.PermissionUsersView
	PermissionUsersCreate        = iamrbac.PermissionUsersCreate
	PermissionUsersDelete        = iamrbac.PermissionUsersDelete
	PermissionUsersResetPassword = iamrbac.PermissionUsersResetPassword
	PermissionUserGroupsView     = iamrbac.PermissionUserGroupsView
	PermissionUserGroupsManage   = iamrbac.PermissionUserGroupsManage
	PermissionRbacView           = iamrbac.PermissionRbacView
	PermissionRbacManage         = iamrbac.PermissionRbacManage
	PermissionSystemSettings     = iamrbac.PermissionSystemSettings
	PermissionSystemLogs         = iamrbac.PermissionSystemLogs
	PermissionSystemImpersonate  = iamrbac.PermissionSystemImpersonate

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

	PermissionChoresManage     = "chores.manage"
	PermissionChoresComplete   = "chores.complete"
	PermissionChoresViewFamily = "chores.view_family"
)

const RoleSuperuser = iamrbac.RoleSuperuser
const RoleMember = iamrbac.RoleMember
const RoleParent = "parent"
const RoleChild = "child"

const GroupEveryone = iamrbac.GroupEveryone
const GroupAdministrators = iamrbac.GroupAdministrators
const GroupParents = "Parents"
const GroupChildren = "Children"

type EffectiveRBAC = iamrbac.EffectiveRBAC
