package server

import (
	"context"
	"testing"

	"connectrpc.com/authn"
	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/auth"
	"github.com/jamesread/starapp/service/internal/rbac"
	"github.com/jamesread/starapp/service/internal/store"
)

func TestUserPreferencesDefaultsAndSave(t *testing.T) {
	st := store.OpenMemory()
	ctx := superuserCtx(context.Background())
	svc := New(nil, st, nil, nil)

	got, err := svc.GetUserPreferences(ctx, connect.NewRequest(&apiv1.GetUserPreferencesRequest{}))
	require.NoError(t, err)
	require.Equal(t, "", got.Msg.Language)
	require.True(t, got.Msg.GetSidebarEnabled())
	require.False(t, got.Msg.GetThemeToggleEnabled())
	require.Contains(t, got.Msg.AvailableLanguages, "en")

	saved, err := svc.SaveUserPreferences(ctx, connect.NewRequest(&apiv1.SaveUserPreferencesRequest{
		Language:           "en",
		SidebarEnabled:     false,
		ThemeToggleEnabled: true,
	}))
	require.NoError(t, err)
	require.True(t, saved.Msg.StandardResponse.Success)

	got2, err := svc.GetUserPreferences(ctx, connect.NewRequest(&apiv1.GetUserPreferencesRequest{}))
	require.NoError(t, err)
	require.Equal(t, "en", got2.Msg.Language)
	require.False(t, got2.Msg.GetSidebarEnabled())
	require.True(t, got2.Msg.GetThemeToggleEnabled())
}

func TestGetStatusIncludesAccountCreated(t *testing.T) {
	st := store.OpenMemory()
	ctx := context.Background()
	id, err := st.CreateUserAccount(ctx, "admin", "hash", store.UserCreatedByAdmin)
	require.NoError(t, err)
	authCtx := authn.SetInfo(ctx, &auth.AuthenticatedUser{
		User: &store.UserAccountRow{ID: id, Username: "admin"},
		RBAC: &rbac.EffectiveRBAC{IsSuperuser: true, Permissions: map[string]bool{}},
	})
	svc := New(nil, st, nil, nil)

	res, err := svc.GetStatus(authCtx, connect.NewRequest(&apiv1.GetStatusRequest{}))
	require.NoError(t, err)
	require.Equal(t, "admin", res.Msg.Username)
	require.NotEmpty(t, res.Msg.AccountCreatedAt)
}
