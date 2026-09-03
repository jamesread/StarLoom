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

func TestUpdateRewardPersistsApprovalRequired(t *testing.T) {
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
	familyID := int(created.Msg.Family.Id)

	rewardID, err := st.CreateReward(ctx, familyID, "Ice cream", "", 5, true, "")
	require.NoError(t, err)

	updated, err := svc.UpdateReward(parentCtx, connect.NewRequest(&apiv1.UpdateRewardRequest{
		Id:               int32(rewardID),
		Title:            "Ice cream",
		CostStars:        5,
		Active:           true,
		ApprovalRequired: false,
	}))
	require.NoError(t, err)
	require.False(t, updated.Msg.Reward.ApprovalRequired)

	reward, err := st.GetRewardByID(ctx, rewardID)
	require.NoError(t, err)
	require.NotNil(t, reward)
	require.False(t, reward.ApprovalRequired)

	updated, err = svc.UpdateReward(parentCtx, connect.NewRequest(&apiv1.UpdateRewardRequest{
		Id:               int32(rewardID),
		Title:            "Ice cream",
		CostStars:        5,
		Active:           true,
		ApprovalRequired: true,
	}))
	require.NoError(t, err)
	require.True(t, updated.Msg.Reward.ApprovalRequired)

	reward, err = st.GetRewardByID(ctx, rewardID)
	require.NoError(t, err)
	require.True(t, reward.ApprovalRequired)
}
