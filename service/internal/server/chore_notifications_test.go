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

func TestChoreNotificationSubscriptionsCRUDAndMatch(t *testing.T) {
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
	parentCtx := userCtx(parentID, "parent", true)
	created, err := svc.CreateFamily(parentCtx, connect.NewRequest(&apiv1.CreateFamilyRequest{Name: "Household"}))
	require.NoError(t, err)
	parentMemberID := int(created.Msg.CallerMember.Id)
	childMemberID, err := st.CreateMember(ctx, int(created.Msg.Family.Id), "Sam", store.MemberRoleChild, &childID, "")
	require.NoError(t, err)

	choreRes, err := svc.CreateChore(parentCtx, connect.NewRequest(&apiv1.CreateChoreRequest{
		Title:          "Make bed",
		StarReward:     1,
		Weekdays:       []int32{1, 2, 3, 4, 5, 6, 7},
		ChildMemberIds: []int32{int32(childMemberID)},
	}))
	require.NoError(t, err)
	choreID := int(choreRes.Msg.Chore.Id)

	_, err = svc.CreateChore(parentCtx, connect.NewRequest(&apiv1.CreateChoreRequest{
		Title:          "Feed cat",
		StarReward:     1,
		Weekdays:       []int32{1, 2, 3, 4, 5, 6, 7},
		ChildMemberIds: []int32{int32(childMemberID)},
	}))
	require.NoError(t, err)

	_, err = svc.SaveMyChoreNotificationSubscriptions(parentCtx, connect.NewRequest(&apiv1.SaveMyChoreNotificationSubscriptionsRequest{
		Subscriptions: []*apiv1.ChoreNotificationSubscription{
			{ChildMemberId: int32(childMemberID), ChoreId: int32(choreID)},
			{ChoreId: int32(choreID)},
		},
	}))
	require.NoError(t, err)

	list, err := svc.GetMyChoreNotificationSubscriptions(parentCtx, connect.NewRequest(&apiv1.GetMyChoreNotificationSubscriptionsRequest{}))
	require.NoError(t, err)
	require.Len(t, list.Msg.Subscriptions, 2)

	match, err := st.MatchingChoreNotificationSubscribers(ctx, int(created.Msg.Family.Id), childMemberID, choreID)
	require.NoError(t, err)
	require.Equal(t, []int{parentMemberID}, match)

	match, err = st.MatchingChoreNotificationSubscribers(ctx, int(created.Msg.Family.Id), parentMemberID, choreID)
	require.NoError(t, err)
	require.Equal(t, []int{parentMemberID}, match)

	_, err = svc.SaveMyChoreNotificationSubscriptions(parentCtx, connect.NewRequest(&apiv1.SaveMyChoreNotificationSubscriptionsRequest{
		Subscriptions: []*apiv1.ChoreNotificationSubscription{
			{ChildMemberId: int32(childMemberID)},
		},
	}))
	require.NoError(t, err)
	list, err = svc.GetMyChoreNotificationSubscriptions(parentCtx, connect.NewRequest(&apiv1.GetMyChoreNotificationSubscriptionsRequest{}))
	require.NoError(t, err)
	require.Len(t, list.Msg.Subscriptions, 1)
	require.Equal(t, int32(childMemberID), list.Msg.Subscriptions[0].ChildMemberId)
	require.Zero(t, list.Msg.Subscriptions[0].ChoreId)
}

func TestChoreNotificationSelfAllChoresSubscription(t *testing.T) {
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

	choreRes, err := svc.CreateChore(parentCtx, connect.NewRequest(&apiv1.CreateChoreRequest{
		Title:          "Make bed",
		StarReward:     1,
		Weekdays:       []int32{1, 2, 3, 4, 5, 6, 7},
		ChildMemberIds: []int32{int32(parentMemberID)},
	}))
	require.NoError(t, err)
	choreID := int(choreRes.Msg.Chore.Id)

	_, err = svc.SaveMyChoreNotificationSubscriptions(parentCtx, connect.NewRequest(&apiv1.SaveMyChoreNotificationSubscriptionsRequest{
		Subscriptions: []*apiv1.ChoreNotificationSubscription{
			{ChildMemberId: int32(parentMemberID)},
		},
	}))
	require.NoError(t, err)

	match, err := st.MatchingChoreNotificationSubscribers(ctx, int(created.Msg.Family.Id), parentMemberID, choreID)
	require.NoError(t, err)
	require.Equal(t, []int{parentMemberID}, match)
}
