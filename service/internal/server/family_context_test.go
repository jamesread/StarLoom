package server

import (
	"context"
	"testing"

	"connectrpc.com/authn"
	"connectrpc.com/connect"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/auth"
	"github.com/jamesread/starapp/service/internal/config"
	"github.com/jamesread/starapp/service/internal/rbac"
	"github.com/jamesread/starapp/service/internal/store"
)

func userCtx(id int, username string, superuser bool) context.Context {
	return authn.SetInfo(context.Background(), &auth.AuthenticatedUser{
		User: &store.UserAccountRow{ID: id, Username: username},
		RBAC: &rbac.EffectiveRBAC{IsSuperuser: superuser, Permissions: map[string]bool{}},
	})
}

func TestSuperuserCanAccessFamilyCreatedByAnotherAdmin(t *testing.T) {
	st := store.OpenMemory()
	t.Cleanup(func() { _ = st.Close() })
	ctx := context.Background()
	require.NoError(t, st.EnsureRBACBootstrap(ctx))
	require.NoError(t, st.SeedDomainRBAC(ctx))

	creatorID, err := st.CreateUserAccount(ctx, "administrator_b", "hash", store.UserCreatedByAdmin)
	require.NoError(t, err)
	adminID, err := st.CreateUserAccount(ctx, "administrator_a", "hash", store.UserCreatedByAdmin)
	require.NoError(t, err)

	svc := New(&config.Config{}, st, nil, logrus.New())
	created, err := svc.CreateFamily(userCtx(creatorID, "administrator_b", true), connect.NewRequest(&apiv1.CreateFamilyRequest{
		Name: "The B Family",
	}))
	require.NoError(t, err)
	require.Equal(t, "The B Family", created.Msg.Family.Name)

	adminCtx := userCtx(adminID, "administrator_a", true)
	home, err := svc.GetParentHomeSummary(adminCtx, connect.NewRequest(&apiv1.GetParentHomeSummaryRequest{}))
	require.NoError(t, err)
	require.Equal(t, created.Msg.Family.Id, home.Msg.Family.Id)
	require.Equal(t, "The B Family", home.Msg.Family.Name)

	mine, err := svc.GetMyFamily(adminCtx, connect.NewRequest(&apiv1.GetMyFamilyRequest{}))
	require.NoError(t, err)
	require.Equal(t, created.Msg.Family.Id, mine.Msg.Family.Id)
	require.NotNil(t, mine.Msg.CallerMember)
	require.Equal(t, store.MemberRoleParent, mine.Msg.CallerMember.Role)
	require.Equal(t, int32(adminID), mine.Msg.CallerMember.UserAccountId)

	members, err := svc.ListMembers(adminCtx, connect.NewRequest(&apiv1.ListMembersRequest{}))
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(members.Msg.Members), 2)
}

func TestUnlinkedChildCannotAccessExistingFamily(t *testing.T) {
	st := store.OpenMemory()
	t.Cleanup(func() { _ = st.Close() })
	ctx := context.Background()
	require.NoError(t, st.EnsureRBACBootstrap(ctx))
	require.NoError(t, st.SeedDomainRBAC(ctx))

	parentID, err := st.CreateUserAccount(ctx, "parent", "hash", store.UserCreatedByAdmin)
	require.NoError(t, err)
	childID, err := st.CreateUserAccount(ctx, "child", "hash", store.UserCreatedByAdmin)
	require.NoError(t, err)

	svc := New(&config.Config{}, st, nil, logrus.New())
	_, err = svc.CreateFamily(userCtx(parentID, "parent", true), connect.NewRequest(&apiv1.CreateFamilyRequest{
		Name: "Household",
	}))
	require.NoError(t, err)

	childPerms := map[string]bool{
		rbac.PermissionAppAccess:          true,
		rbac.PermissionStarsViewOwn:       true,
		rbac.PermissionRewardsView:        true,
		rbac.PermissionRedemptionsRequest: true,
		rbac.PermissionChoresViewFamily:   true,
	}
	childCtx := authn.SetInfo(context.Background(), &auth.AuthenticatedUser{
		User: &store.UserAccountRow{ID: childID, Username: "child"},
		RBAC: &rbac.EffectiveRBAC{Permissions: childPerms},
	})
	_, err = svc.GetParentHomeSummary(childCtx, connect.NewRequest(&apiv1.GetParentHomeSummaryRequest{}))
	require.Error(t, err)
	require.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))

	member, err := st.GetMemberByAccountID(ctx, childID)
	require.NoError(t, err)
	require.Nil(t, member)
}
