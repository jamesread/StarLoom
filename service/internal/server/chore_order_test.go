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

func TestReorderChoresDrivesListAndChartOrder(t *testing.T) {
	st := store.OpenMemory()
	t.Cleanup(func() { _ = st.Close() })
	ctx := context.Background()
	require.NoError(t, st.EnsureRBACBootstrap(ctx))
	require.NoError(t, st.SeedDomainRBAC(ctx))

	parentID, err := st.CreateUserAccount(ctx, "parent", "hash", store.UserCreatedByAdmin)
	require.NoError(t, err)
	parentCtx := userCtx(parentID, "parent", true)

	svc := New(&config.Config{}, st, nil, logrus.New())
	created, err := svc.CreateFamily(parentCtx, connect.NewRequest(&apiv1.CreateFamilyRequest{Name: "Household"}))
	require.NoError(t, err)
	familyID := int(created.Msg.Family.Id)

	childMemberID, err := st.CreateMember(ctx, familyID, "Sam", store.MemberRoleChild, nil, "")
	require.NoError(t, err)

	chart, err := svc.CreateStarChart(parentCtx, connect.NewRequest(&apiv1.CreateStarChartRequest{Name: "Chores"}))
	require.NoError(t, err)
	chartID := chart.Msg.StarChart.Id

	ids := map[string]int32{}
	for _, title := range []string{"Apples", "Bananas", "Cherries"} {
		chore, err := svc.CreateChore(parentCtx, connect.NewRequest(&apiv1.CreateChoreRequest{
			Title:          title,
			StarReward:     1,
			Weekdays:       []int32{1, 2, 3, 4, 5, 6, 7},
			ChildMemberIds: []int32{int32(childMemberID)},
			StarChartId:    chartID,
		}))
		require.NoError(t, err)
		ids[title] = chore.Msg.Chore.Id
	}

	_, err = svc.ReorderChores(parentCtx, connect.NewRequest(&apiv1.ReorderChoresRequest{
		ChoreIds: []int32{ids["Cherries"], ids["Apples"], ids["Bananas"]},
	}))
	require.NoError(t, err)

	listed, err := svc.ListChores(parentCtx, connect.NewRequest(&apiv1.ListChoresRequest{StarChartId: chartID}))
	require.NoError(t, err)
	titles := make([]string, 0, len(listed.Msg.Chores))
	for _, c := range listed.Msg.Chores {
		titles = append(titles, c.Title)
	}
	require.Equal(t, []string{"Cherries", "Apples", "Bananas"}, titles)

	weekly, err := svc.GetWeeklyStarChart(parentCtx, connect.NewRequest(&apiv1.GetWeeklyStarChartRequest{
		StarChartId: chartID,
	}))
	require.NoError(t, err)
	rowTitles := make([]string, 0, len(weekly.Msg.Rows))
	for _, row := range weekly.Msg.Rows {
		rowTitles = append(rowTitles, row.Title)
	}
	require.Equal(t, []string{"Cherries", "Apples", "Bananas"}, rowTitles)
}

func TestReorderChoresRejectsBadIDs(t *testing.T) {
	st := store.OpenMemory()
	t.Cleanup(func() { _ = st.Close() })
	ctx := context.Background()
	require.NoError(t, st.EnsureRBACBootstrap(ctx))
	require.NoError(t, st.SeedDomainRBAC(ctx))

	parentID, err := st.CreateUserAccount(ctx, "parent", "hash", store.UserCreatedByAdmin)
	require.NoError(t, err)
	parentCtx := userCtx(parentID, "parent", true)
	svc := New(&config.Config{}, st, nil, logrus.New())
	created, err := svc.CreateFamily(parentCtx, connect.NewRequest(&apiv1.CreateFamilyRequest{Name: "Household"}))
	require.NoError(t, err)
	childMemberID, err := st.CreateMember(ctx, int(created.Msg.Family.Id), "Sam", store.MemberRoleChild, nil, "")
	require.NoError(t, err)
	chart, err := svc.CreateStarChart(parentCtx, connect.NewRequest(&apiv1.CreateStarChartRequest{Name: "Chores"}))
	require.NoError(t, err)
	chartID := chart.Msg.StarChart.Id

	ids := make([]int32, 0, 2)
	for _, title := range []string{"Apples", "Bananas"} {
		chore, err := svc.CreateChore(parentCtx, connect.NewRequest(&apiv1.CreateChoreRequest{
			Title:          title,
			StarReward:     1,
			Weekdays:       []int32{1, 2, 3, 4, 5, 6, 7},
			ChildMemberIds: []int32{int32(childMemberID)},
			StarChartId:    chartID,
		}))
		require.NoError(t, err)
		ids = append(ids, chore.Msg.Chore.Id)
	}

	// Second family, store level only.
	otherFamilyID, err := st.CreateFamily(ctx, "Neighbours")
	require.NoError(t, err)
	foreignChoreID, err := st.CreateChore(ctx, otherFamilyID, 0, "Not ours", 1, 127, nil)
	require.NoError(t, err)

	for name, choreIDs := range map[string][]int32{
		"empty":     {},
		"duplicate": {ids[0], ids[0]},
		"unknown":   {ids[0], 9999},
		"foreign":   {ids[0], int32(foreignChoreID)},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := svc.ReorderChores(parentCtx, connect.NewRequest(&apiv1.ReorderChoresRequest{
				ChoreIds: choreIDs,
			}))
			require.Error(t, err)
			require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
		})
	}

	listed, err := svc.ListChores(parentCtx, connect.NewRequest(&apiv1.ListChoresRequest{StarChartId: chartID}))
	require.NoError(t, err)
	titles := make([]string, 0, len(listed.Msg.Chores))
	for _, c := range listed.Msg.Chores {
		titles = append(titles, c.Title)
	}
	require.Equal(t, []string{"Apples", "Bananas"}, titles)
}
