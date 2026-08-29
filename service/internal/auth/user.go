package auth

import (
	"github.com/jamesread/starapp/service/internal/rbac"
	"github.com/jamesread/starapp/service/internal/store"
)

type AuthenticatedUser struct {
	User     *store.UserAccountRow
	RBAC     *rbac.EffectiveRBAC
	ReadOnly bool
}

func (a *AuthenticatedUser) HasPermission(p string) bool {
	if a == nil || a.RBAC == nil {
		return false
	}
	return a.RBAC.Has(p)
}

func (a *AuthenticatedUser) hasAnyPermission(perms ...string) bool {
	for _, p := range perms {
		if a.HasPermission(p) {
			return true
		}
	}
	return false
}

func (a *AuthenticatedUser) CanAccessIam() bool {
	return a.hasAnyPermission(
		rbac.PermissionUsersView,
		rbac.PermissionUserGroupsView,
		rbac.PermissionRbacView,
	)
}

func (a *AuthenticatedUser) CanAccessSettings() bool {
	return a.HasPermission(rbac.PermissionSystemSettings)
}

func (a *AuthenticatedUser) CanAccessWebhooks() bool {
	return a.CanAccessSettings()
}

func (a *AuthenticatedUser) CanAccessControlPanel() bool {
	return a.CanAccessIam() || a.CanAccessSettings()
}
