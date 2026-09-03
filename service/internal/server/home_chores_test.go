package server

import (
	"context"
	"testing"
	"time"

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

func childPerms() map[string]bool {
	return map[string]bool{
		rbac.PermissionAppAccess:        true,
		rbac.PermissionStarsViewOwn:     true,
		rbac.PermissionRewardsView:      true,
		rbac.PermissionChoresViewFamily: true,
	}
}

func TestChildHomeListsAndCompletesOwnTodaysChore(t *testing.T) {
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
	created, err := svc.CreateFamily(userCtx(parentID, "parent", true), connect.NewRequest(&apiv1.CreateFamilyRequest{
		Name: "Household",
	}))
	require.NoError(t, err)
	familyID := int(created.Msg.Family.Id)

	childMemberID, err := st.CreateMember(ctx, familyID, "Sam", store.MemberRoleChild, &childID, "")
	require.NoError(t, err)
	createdChore, err := svc.CreateChore(userCtx(parentID, "parent", true), connect.NewRequest(&apiv1.CreateChoreRequest{
		Title:          "Make the bed",
		StarReward:     2,
		Weekdays:       []int32{1, 2, 3, 4, 5, 6, 7},
		ChildMemberIds: []int32{int32(childMemberID)},
	}))
	require.NoError(t, err)
	choreID := int(createdChore.Msg.Chore.Id)

	childCtx := authn.SetInfo(context.Background(), &auth.AuthenticatedUser{
		User: &store.UserAccountRow{ID: childID, Username: "child"},
		RBAC: &rbac.EffectiveRBAC{Permissions: childPerms()},
	})

	home, err := svc.GetChildHomeSummary(childCtx, connect.NewRequest(&apiv1.GetChildHomeSummaryRequest{}))
	require.NoError(t, err)
	require.Len(t, home.Msg.TodaysChores, 1)
	require.Equal(t, "Make the bed", home.Msg.TodaysChores[0].Title)
	require.False(t, home.Msg.TodaysChores[0].Completed)
	require.Equal(t, int32(childMemberID), home.Msg.TodaysChores[0].ChildMemberId)

	done, err := svc.CompleteChore(childCtx, connect.NewRequest(&apiv1.CompleteChoreRequest{
		ChoreId:       int32(choreID),
		ChildMemberId: int32(childMemberID),
	}))
	require.NoError(t, err)
	require.True(t, done.Msg.StandardResponse.Success)
	require.Equal(t, int32(2), done.Msg.NewBalance)

	home, err = svc.GetChildHomeSummary(childCtx, connect.NewRequest(&apiv1.GetChildHomeSummaryRequest{}))
	require.NoError(t, err)
	require.True(t, home.Msg.TodaysChores[0].Completed)
}

func TestChildCannotCompleteSiblingChore(t *testing.T) {
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
	created, err := svc.CreateFamily(userCtx(parentID, "parent", true), connect.NewRequest(&apiv1.CreateFamilyRequest{
		Name: "Household",
	}))
	require.NoError(t, err)
	familyID := int(created.Msg.Family.Id)
	childMemberID, err := st.CreateMember(ctx, familyID, "Sam", store.MemberRoleChild, &childID, "")
	require.NoError(t, err)
	siblingID, err := st.CreateMember(ctx, familyID, "Alex", store.MemberRoleChild, nil, "")
	require.NoError(t, err)
	createdChore, err := svc.CreateChore(userCtx(parentID, "parent", true), connect.NewRequest(&apiv1.CreateChoreRequest{
		Title:          "Tidy room",
		StarReward:     1,
		Weekdays:       []int32{1, 2, 3, 4, 5, 6, 7},
		ChildMemberIds: []int32{int32(siblingID)},
	}))
	require.NoError(t, err)
	choreID := int(createdChore.Msg.Chore.Id)

	childCtx := authn.SetInfo(context.Background(), &auth.AuthenticatedUser{
		User: &store.UserAccountRow{ID: childID, Username: "child"},
		RBAC: &rbac.EffectiveRBAC{Permissions: childPerms()},
	})
	_, err = svc.CompleteChore(childCtx, connect.NewRequest(&apiv1.CompleteChoreRequest{
		ChoreId:       int32(choreID),
		ChildMemberId: int32(siblingID),
	}))
	require.Error(t, err)
	require.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))

	home, err := svc.GetChildHomeSummary(childCtx, connect.NewRequest(&apiv1.GetChildHomeSummaryRequest{}))
	require.NoError(t, err)
	require.Empty(t, home.Msg.TodaysChores)
	_ = childMemberID
}

func TestParentHomeListsOnlyOwnTodaysChores(t *testing.T) {
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
	childMemberID, err := st.CreateMember(ctx, int(created.Msg.Family.Id), "Sam", store.MemberRoleChild, nil, "")
	require.NoError(t, err)
	_, err = svc.CreateChore(parentCtx, connect.NewRequest(&apiv1.CreateChoreRequest{
		Title:          "Feed the cat",
		StarReward:     1,
		Weekdays:       []int32{1, 2, 3, 4, 5, 6, 7},
		ChildMemberIds: []int32{int32(parentMemberID)},
	}))
	require.NoError(t, err)
	_, err = svc.CreateChore(parentCtx, connect.NewRequest(&apiv1.CreateChoreRequest{
		Title:          "Sam only",
		StarReward:     1,
		Weekdays:       []int32{1, 2, 3, 4, 5, 6, 7},
		ChildMemberIds: []int32{int32(childMemberID)},
	}))
	require.NoError(t, err)

	home, err := svc.GetParentHomeSummary(parentCtx, connect.NewRequest(&apiv1.GetParentHomeSummaryRequest{}))
	require.NoError(t, err)
	require.Len(t, home.Msg.TodaysChores, 1)
	require.Equal(t, "Feed the cat", home.Msg.TodaysChores[0].Title)
	require.Equal(t, int32(parentMemberID), home.Msg.TodaysChores[0].ChildMemberId)
}

func TestParentHomeOmitsChoresNotScheduledToday(t *testing.T) {
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
	today := weekdayFromToday()
	other := today%7 + 1
	_, err = svc.CreateChore(parentCtx, connect.NewRequest(&apiv1.CreateChoreRequest{
		Title:          "Only another day",
		StarReward:     1,
		Weekdays:       []int32{int32(other)},
		ChildMemberIds: []int32{int32(parentMemberID)},
	}))
	require.NoError(t, err)

	home, err := svc.GetParentHomeSummary(parentCtx, connect.NewRequest(&apiv1.GetParentHomeSummaryRequest{}))
	require.NoError(t, err)
	require.Empty(t, home.Msg.TodaysChores)
}

func weekdayFromToday() int32 {
	// Monday=1 … Sunday=7, matching chore weekday masks.
	wd := time.Now().Weekday()
	if wd == time.Sunday {
		return 7
	}
	return int32(wd)
}

func TestGetMemberTodaysChoresForChild(t *testing.T) {
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
	childMemberID, err := st.CreateMember(ctx, int(created.Msg.Family.Id), "Sam", store.MemberRoleChild, nil, "")
	require.NoError(t, err)
	choreRes, err := svc.CreateChore(parentCtx, connect.NewRequest(&apiv1.CreateChoreRequest{
		Title:          "Make bed",
		StarReward:     2,
		Weekdays:       []int32{1, 2, 3, 4, 5, 6, 7},
		ChildMemberIds: []int32{int32(childMemberID)},
	}))
	require.NoError(t, err)
	choreID := int(choreRes.Msg.Chore.Id)

	got, err := svc.GetMemberTodaysChores(parentCtx, connect.NewRequest(&apiv1.GetMemberTodaysChoresRequest{
		MemberId: int32(childMemberID),
	}))
	require.NoError(t, err)
	require.Len(t, got.Msg.TodaysChores, 1)
	require.Equal(t, "Make bed", got.Msg.TodaysChores[0].Title)
	require.Equal(t, int32(childMemberID), got.Msg.TodaysChores[0].ChildMemberId)

	done, err := svc.CompleteChore(parentCtx, connect.NewRequest(&apiv1.CompleteChoreRequest{
		ChoreId:       int32(choreID),
		ChildMemberId: int32(childMemberID),
	}))
	require.NoError(t, err)
	require.True(t, done.Msg.StandardResponse.Success)

	got, err = svc.GetMemberTodaysChores(parentCtx, connect.NewRequest(&apiv1.GetMemberTodaysChoresRequest{
		MemberId: int32(childMemberID),
	}))
	require.NoError(t, err)
	require.True(t, got.Msg.TodaysChores[0].Completed)

	own, err := svc.GetMemberTodaysChores(parentCtx, connect.NewRequest(&apiv1.GetMemberTodaysChoresRequest{
		MemberId: int32(parentMemberID),
	}))
	require.NoError(t, err)
	require.Empty(t, own.Msg.TodaysChores)
}

func TestParentLinkedMemberGetsOwnHomeRewards(t *testing.T) {
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

	_, err = svc.CreateReward(parentCtx, connect.NewRequest(&apiv1.CreateRewardRequest{
		Title:     "Extra screen time",
		CostStars: 5,
	}))
	require.NoError(t, err)
	_, err = st.InsertLedgerEntry(ctx, store.StarLedgerRow{
		FamilyID: int(created.Msg.Family.Id), ChildMemberID: parentMemberID, Amount: 10,
		EntryType: store.LedgerTypeAward, Note: "Good week",
	})
	require.NoError(t, err)

	home, err := svc.GetChildHomeSummary(parentCtx, connect.NewRequest(&apiv1.GetChildHomeSummaryRequest{}))
	require.NoError(t, err)
	require.Equal(t, int32(parentMemberID), home.Msg.Member.Id)
	require.Equal(t, int32(10), home.Msg.Balance)
	require.Len(t, home.Msg.Rewards, 1)
	require.Equal(t, "Extra screen time", home.Msg.Rewards[0].Title)
}

func TestParentHomeIncludesChildTodayStarChartProgress(t *testing.T) {
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

	charts, err := svc.ListStarCharts(parentCtx, connect.NewRequest(&apiv1.ListStarChartsRequest{}))
	require.NoError(t, err)
	require.NotEmpty(t, charts.Msg.StarCharts)
	chartID := charts.Msg.StarCharts[0].Id
	chartName := charts.Msg.StarCharts[0].Name

	choreOne, err := svc.CreateChore(parentCtx, connect.NewRequest(&apiv1.CreateChoreRequest{
		Title:          "Make bed",
		StarReward:     1,
		Weekdays:       []int32{1, 2, 3, 4, 5, 6, 7},
		ChildMemberIds: []int32{int32(childMemberID)},
		StarChartId:    chartID,
	}))
	require.NoError(t, err)
	_, err = svc.CreateChore(parentCtx, connect.NewRequest(&apiv1.CreateChoreRequest{
		Title:          "Feed cat",
		StarReward:     1,
		Weekdays:       []int32{1, 2, 3, 4, 5, 6, 7},
		ChildMemberIds: []int32{int32(childMemberID)},
		StarChartId:    chartID,
	}))
	require.NoError(t, err)

	done, err := svc.CompleteChore(parentCtx, connect.NewRequest(&apiv1.CompleteChoreRequest{
		ChoreId:       choreOne.Msg.Chore.Id,
		ChildMemberId: int32(childMemberID),
	}))
	require.NoError(t, err)
	require.True(t, done.Msg.StandardResponse.Success)

	home, err := svc.GetParentHomeSummary(parentCtx, connect.NewRequest(&apiv1.GetParentHomeSummaryRequest{}))
	require.NoError(t, err)
	var childSummary *apiv1.ChildHomeSummary
	for _, entry := range home.Msg.Children {
		if entry.GetMember().GetId() == int32(childMemberID) {
			childSummary = entry
			break
		}
	}
	require.NotNil(t, childSummary)
	require.Len(t, childSummary.TodayStarChartProgress, 1)
	require.Equal(t, chartID, childSummary.TodayStarChartProgress[0].StarChartId)
	require.Equal(t, chartName, childSummary.TodayStarChartProgress[0].StarChartName)
	require.Equal(t, int32(1), childSummary.TodayStarChartProgress[0].Completed)
	require.Equal(t, int32(2), childSummary.TodayStarChartProgress[0].Scheduled)
	require.False(t, childSummary.TodayStarChartProgress[0].Paused)
}
