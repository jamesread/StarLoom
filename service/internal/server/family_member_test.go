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
