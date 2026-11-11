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

	rewardID, err := st.CreateReward(ctx, familyID, "Ice cream", "", 5, true, "", false)
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

func TestRequestRedemptionQuantity(t *testing.T) {
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
	childID, err := st.CreateMember(ctx, familyID, "Sam", store.MemberRoleChild, nil, "")
	require.NoError(t, err)
	_, err = svc.AwardStars(parentCtx, connect.NewRequest(&apiv1.AwardStarsRequest{ChildMemberId: int32(childID), Amount: 20}))
	require.NoError(t, err)

	multiID, err := st.CreateReward(ctx, familyID, "Pocket money", "", 2, false, "", true)
	require.NoError(t, err)
	singleID, err := st.CreateReward(ctx, familyID, "Ice cream", "", 2, false, "", false)
	require.NoError(t, err)

	res, err := svc.RequestRedemption(parentCtx, connect.NewRequest(&apiv1.RequestRedemptionRequest{
		RewardId: int32(multiID), ChildMemberId: int32(childID), Quantity: 5,
	}))
	require.NoError(t, err)
	require.Equal(t, int32(10), res.Msg.Redemption.StarsSpent)
	balance, err := st.GetMemberBalance(ctx, childID)
	require.NoError(t, err)
	require.Equal(t, 10, balance)

	_, err = svc.RequestRedemption(parentCtx, connect.NewRequest(&apiv1.RequestRedemptionRequest{
		RewardId: int32(singleID), ChildMemberId: int32(childID), Quantity: 2,
	}))
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))

	_, err = svc.RequestRedemption(parentCtx, connect.NewRequest(&apiv1.RequestRedemptionRequest{
		RewardId: int32(multiID), ChildMemberId: int32(childID), Quantity: 6,
	}))
	require.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
}
