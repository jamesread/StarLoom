package server

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/config"
	"github.com/jamesread/starapp/service/internal/store"
)

func TestUpdateMemberAppliesStarAdjustment(t *testing.T) {
	st := store.OpenMemory()
	t.Cleanup(func() { _ = st.Close() })
	ctx := context.Background()
	require.NoError(t, st.EnsureRBACBootstrap(ctx))
	require.NoError(t, st.SeedDomainRBAC(ctx))

	parentID, err := st.CreateUserAccount(ctx, "parent", "hash", store.UserCreatedByAdmin)
	require.NoError(t, err)

	svc := New(&config.Config{}, st, nil, logrus.New())
	created, err := svc.CreateFamily(userCtx(parentID, "parent", true), connect.NewRequest(&apiv1.CreateFamilyRequest{
		Name: "Household",
	}))
	require.NoError(t, err)
	familyID := int(created.Msg.Family.Id)

	childMemberID, err := st.CreateMember(ctx, familyID, "Sam", store.MemberRoleChild, nil, "")
	require.NoError(t, err)

	parentCtx := userCtx(parentID, "parent", true)

	updated, err := svc.UpdateMember(parentCtx, connect.NewRequest(&apiv1.UpdateMemberRequest{
		MemberId:       int32(childMemberID),
		DisplayName:    "Sam",
		StarAdjustment: 5,
		AdjustmentNote: "Welcome bonus",
	}))
	require.NoError(t, err)
	require.Equal(t, int32(5), updated.Msg.NewBalance)

	bal, err := svc.GetMemberBalance(parentCtx, connect.NewRequest(&apiv1.GetMemberBalanceRequest{
		MemberId: int32(childMemberID),
	}))
	require.NoError(t, err)
	require.Equal(t, int32(5), bal.Msg.Balance)

	updated, err = svc.UpdateMember(parentCtx, connect.NewRequest(&apiv1.UpdateMemberRequest{
		MemberId:       int32(childMemberID),
		DisplayName:    "Sam",
		StarAdjustment: -2,
	}))
	require.NoError(t, err)
	require.Equal(t, int32(3), updated.Msg.NewBalance)
}

func TestDeleteMemberRemovesChild(t *testing.T) {
	st := store.OpenMemory()
	t.Cleanup(func() { _ = st.Close() })
	ctx := context.Background()
	require.NoError(t, st.EnsureRBACBootstrap(ctx))
	require.NoError(t, st.SeedDomainRBAC(ctx))

	parentID, err := st.CreateUserAccount(ctx, "parent", "hash", store.UserCreatedByAdmin)
	require.NoError(t, err)

	svc := New(&config.Config{}, st, nil, logrus.New())
	parentCtx := userCtx(parentID, "parent", true)
	created, err := svc.CreateFamily(parentCtx, connect.NewRequest(&apiv1.CreateFamilyRequest{Name: "Household"}))
	require.NoError(t, err)
	childMemberID, err := st.CreateMember(ctx, int(created.Msg.Family.Id), "Sam", store.MemberRoleChild, nil, "")
	require.NoError(t, err)

	_, err = svc.DeleteMember(parentCtx, connect.NewRequest(&apiv1.DeleteMemberRequest{MemberId: int32(childMemberID)}))
	require.NoError(t, err)

	members, err := svc.ListMembers(parentCtx, connect.NewRequest(&apiv1.ListMembersRequest{}))
	require.NoError(t, err)
	for _, m := range members.Msg.Members {
		require.NotEqual(t, int32(childMemberID), m.Id)
	}
}

func TestDeleteMemberRemovesParentWhenAnotherParentExists(t *testing.T) {
	st := store.OpenMemory()
	t.Cleanup(func() { _ = st.Close() })
	ctx := context.Background()
	require.NoError(t, st.EnsureRBACBootstrap(ctx))
	require.NoError(t, st.SeedDomainRBAC(ctx))

	parentAID, err := st.CreateUserAccount(ctx, "parent-a", "hash", store.UserCreatedByAdmin)
	require.NoError(t, err)
	parentBID, err := st.CreateUserAccount(ctx, "parent-b", "hash", store.UserCreatedByAdmin)
	require.NoError(t, err)

	svc := New(&config.Config{}, st, nil, logrus.New())
	parentACtx := userCtx(parentAID, "parent-a", true)
	created, err := svc.CreateFamily(parentACtx, connect.NewRequest(&apiv1.CreateFamilyRequest{Name: "Household"}))
	require.NoError(t, err)
	familyID := int(created.Msg.Family.Id)
	parentBMemberID, err := st.CreateMember(ctx, familyID, "Parent B", store.MemberRoleParent, &parentBID, "")
	require.NoError(t, err)

	_, err = svc.DeleteMember(parentACtx, connect.NewRequest(&apiv1.DeleteMemberRequest{MemberId: int32(parentBMemberID)}))
	require.NoError(t, err)

	remaining, err := st.GetUserByID(ctx, parentBID)
	require.NoError(t, err)
	require.NotNil(t, remaining)
}

func TestDeleteMemberRejectsSelfRemoval(t *testing.T) {
	st := store.OpenMemory()
	t.Cleanup(func() { _ = st.Close() })
	ctx := context.Background()
	require.NoError(t, st.EnsureRBACBootstrap(ctx))
	require.NoError(t, st.SeedDomainRBAC(ctx))

	parentID, err := st.CreateUserAccount(ctx, "parent", "hash", store.UserCreatedByAdmin)
	require.NoError(t, err)

	svc := New(&config.Config{}, st, nil, logrus.New())
	parentCtx := userCtx(parentID, "parent", true)
	created, err := svc.CreateFamily(parentCtx, connect.NewRequest(&apiv1.CreateFamilyRequest{Name: "Household"}))
	require.NoError(t, err)
	parentMemberID := int(created.Msg.CallerMember.Id)

	_, err = svc.DeleteMember(parentCtx, connect.NewRequest(&apiv1.DeleteMemberRequest{MemberId: int32(parentMemberID)}))
	require.Error(t, err)
	require.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
}
