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

func TestListStarChartsAssignedToMe(t *testing.T) {
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

	charts, err := svc.ListStarCharts(parentCtx, connect.NewRequest(&apiv1.ListStarChartsRequest{}))
	require.NoError(t, err)
	require.NotEmpty(t, charts.Msg.StarCharts)
	chartID := charts.Msg.StarCharts[0].Id

	_, err = svc.CreateChore(parentCtx, connect.NewRequest(&apiv1.CreateChoreRequest{
		Title:          "Child chore",
		StarReward:     1,
		Weekdays:       []int32{1, 2, 3, 4, 5, 6, 7},
		ChildMemberIds: []int32{int32(childMemberID)},
		StarChartId:    chartID,
	}))
	require.NoError(t, err)

	mine, err := svc.ListStarCharts(parentCtx, connect.NewRequest(&apiv1.ListStarChartsRequest{AssignedToMe: true}))
	require.NoError(t, err)
	require.Empty(t, mine.Msg.StarCharts)

	_, err = svc.CreateChore(parentCtx, connect.NewRequest(&apiv1.CreateChoreRequest{
		Title:          "Parent chore",
		StarReward:     1,
		Weekdays:       []int32{1, 2, 3, 4, 5, 6, 7},
		ChildMemberIds: []int32{int32(parentMemberID)},
		StarChartId:    chartID,
	}))
	require.NoError(t, err)

	mine, err = svc.ListStarCharts(parentCtx, connect.NewRequest(&apiv1.ListStarChartsRequest{AssignedToMe: true}))
	require.NoError(t, err)
	require.Len(t, mine.Msg.StarCharts, 1)
	require.Equal(t, int32(1), mine.Msg.StarCharts[0].ChoreCount)

	childCtx := authn.SetInfo(context.Background(), &auth.AuthenticatedUser{
		User: &store.UserAccountRow{ID: childID, Username: "child"},
		RBAC: &rbac.EffectiveRBAC{Permissions: childPerms()},
	})
	childCharts, err := svc.ListStarCharts(childCtx, connect.NewRequest(&apiv1.ListStarChartsRequest{AssignedToMe: true}))
	require.NoError(t, err)
	require.Len(t, childCharts.Msg.StarCharts, 1)
	require.Equal(t, int32(1), childCharts.Msg.StarCharts[0].ChoreCount)
}
