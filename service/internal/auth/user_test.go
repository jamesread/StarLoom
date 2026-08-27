package auth

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jamesread/starapp/service/internal/rbac"
)

func userWith(perms ...string) *AuthenticatedUser {
	set := map[string]bool{}
	for _, p := range perms {
		set[p] = true
	}
	return &AuthenticatedUser{RBAC: &rbac.EffectiveRBAC{Permissions: set}}
}

func TestCanAccessControlPanel(t *testing.T) {
	require.False(t, (*AuthenticatedUser)(nil).CanAccessControlPanel())
	require.False(t, userWith().CanAccessControlPanel())
	require.False(t, userWith(rbac.PermissionSystemLogs, rbac.PermissionSystemImpersonate).CanAccessControlPanel())

	require.True(t, (&AuthenticatedUser{RBAC: &rbac.EffectiveRBAC{IsSuperuser: true}}).CanAccessControlPanel())
	require.True(t, userWith(rbac.PermissionUsersView).CanAccessControlPanel())
	require.True(t, userWith(rbac.PermissionSystemSettings).CanAccessControlPanel())
}

func TestCanAccessIamAndSettings(t *testing.T) {
	iamOnly := userWith(rbac.PermissionRbacView)
	require.True(t, iamOnly.CanAccessIam())
	require.False(t, iamOnly.CanAccessSettings())
	require.True(t, iamOnly.CanAccessControlPanel())

	settingsOnly := userWith(rbac.PermissionSystemSettings)
	require.False(t, settingsOnly.CanAccessIam())
	require.True(t, settingsOnly.CanAccessSettings())
	require.True(t, settingsOnly.CanAccessControlPanel())
}
