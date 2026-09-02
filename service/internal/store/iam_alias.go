package store

import iamstore "github.com/jamesread/armature-iam/store"

const (
	UserCreatedByAdmin = iamstore.UserCreatedByAdmin
	UserCreatedBySSO   = iamstore.UserCreatedBySSO
)

type (
	UserAccountRow       = iamstore.UserAccountRow
	SessionRow           = iamstore.SessionRow
	APIKeyRow            = iamstore.APIKeyRow
	UserGroupRow         = iamstore.UserGroupRow
	RBACPermissionRow    = iamstore.RBACPermissionRow
	RBACRoleRow          = iamstore.RBACRoleRow
	MyPermissionAuditRow = iamstore.MyPermissionAuditRow
)
