package server

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/config"
	"github.com/jamesread/starapp/service/internal/store"
)

func TestRewardAvailabilityUsesRedemptionCounts(t *testing.T) {
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
	childMemberID, err := st.CreateMember(ctx, familyID, "Sam", store.MemberRoleChild, nil, "")
	require.NoError(t, err)

	rewardRes, err := svc.CreateReward(parentCtx, connect.NewRequest(&apiv1.CreateRewardRequest{
		Title:                  "Daily treat",
		CostStars:              1,
		AvailabilityExpression: "countPerDay < 1",
	}))
	require.NoError(t, err)
	rewardID := int(rewardRes.Msg.Reward.Id)

	_, err = st.CreateRedemption(ctx, familyID, childMemberID, rewardID, 1, store.RedemptionApproved, nil)
	require.NoError(t, err)

	reward, err := st.GetRewardByID(ctx, rewardID)
	require.NoError(t, err)
	require.NotNil(t, reward)
	now := time.Now()
	require.False(t, svc.rewardAvailableNow(ctx, reward, childMemberID, 9, now))

	env, err := svc.rewardAvailabilityEnv(ctx, reward, childMemberID, 9, now)
	require.NoError(t, err)
	require.Equal(t, 1, env.CountPerDay)
	require.Equal(t, 1, env.CountPerWeek)
}
