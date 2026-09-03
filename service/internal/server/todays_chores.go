package server

import (
	"context"
	"fmt"
	"sort"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/rbac"
	"github.com/jamesread/starapp/service/internal/store"
)

func (s *Server) listTodaysChores(ctx context.Context, fc *familyContext, memberFilter int) ([]*apiv1.TodaysChore, error) {
	date := time.Now().Format("2006-01-02")
	paused, err := s.store.IsDatePaused(ctx, fc.family.ID, date)
	if err != nil {
		return nil, err
	}
	chores, err := s.store.ListChores(ctx, fc.family.ID, 0, false)
	if err != nil {
		return nil, err
	}
	done, err := s.todaysCompletionSet(ctx, fc.family.ID, date)
	if err != nil {
		return nil, err
	}
	members, err := s.store.ListMembersByFamily(ctx, fc.family.ID)
	if err != nil {
		return nil, err
	}
	memberByID := map[int]store.FamilyMemberRow{}
	for _, m := range members {
		memberByID[m.ID] = m
	}
	chartRows, err := s.store.ListStarCharts(ctx, fc.family.ID, false)
	if err != nil {
		return nil, err
	}
	chartNames := map[int]string{}
	for i := range chartRows {
		chartNames[chartRows[i].ID] = chartRows[i].Name
	}
	defaultChartID, err := s.store.GetDefaultStarChartID(ctx, fc.family.ID)
	if err != nil {
		return nil, err
	}
	out := make([]*apiv1.TodaysChore, 0)
	for i := range chores {
		out = appendTodaysAssignments(&chores[i], memberFilter, date, paused, done, memberByID, chartNames, defaultChartID, out)
	}
	sortTodaysChores(out)
	return out, nil
}

func (s *Server) todaysCompletionSet(ctx context.Context, familyID int, date string) (map[string]bool, error) {
	rows, err := s.store.ListCompletionsForWeek(ctx, familyID, date, date)
	if err != nil {
		return nil, err
	}
	done := map[string]bool{}
	for _, row := range rows {
		done[fmt.Sprintf("%d:%d", row.ChoreID, row.ChildMemberID)] = true
	}
	return done, nil
}

func appendTodaysAssignments(
	cw *store.ChoreWithAssignments,
	memberFilter int,
	date string,
	paused bool,
	done map[string]bool,
	memberByID map[int]store.FamilyMemberRow,
	chartNames map[int]string,
	defaultChartID int,
	out []*apiv1.TodaysChore,
) []*apiv1.TodaysChore {
	if cw == nil || !store.IsScheduledOnDate(cw.Chore.WeekdayMask, date) {
		return out
	}
	chartID := cw.Chore.StarChartID
	if chartID == 0 {
		chartID = defaultChartID
	}
	chartName := chartNames[chartID]
	if chartName == "" {
		chartName = "Star Chart"
	}
	for _, assign := range cw.Assignments {
		if memberFilter != 0 && assign.ChildMemberID != memberFilter {
			continue
		}
		member := memberByID[assign.ChildMemberID]
		out = append(out, &apiv1.TodaysChore{
			ChoreId:       int32(cw.Chore.ID),
			Title:         cw.Chore.Title,
			StarReward:    int32(cw.Chore.StarReward),
			ChildMemberId: int32(assign.ChildMemberID),
			Child:         toProtoMember(&member),
			Completed:     done[fmt.Sprintf("%d:%d", cw.Chore.ID, assign.ChildMemberID)],
			Paused:        paused,
			Date:          date,
			StarChartId:   int32(chartID),
			StarChartName: chartName,
		})
	}
	return out
}

func sortTodaysChores(rows []*apiv1.TodaysChore) {
	sort.Slice(rows, func(i, j int) bool {
		leftChart := rows[i].GetStarChartName()
		rightChart := rows[j].GetStarChartName()
		if leftChart != rightChart {
			return leftChart < rightChart
		}
		left := rows[i].GetChild().GetDisplayName()
		right := rows[j].GetChild().GetDisplayName()
		if left != right {
			return left < right
		}
		return rows[i].Title < rows[j].Title
	})
}

func (s *Server) memberTodayStarChartProgress(ctx context.Context, fc *familyContext, memberID int) ([]*apiv1.StarChartDayProgress, error) {
	date := time.Now().Format("2006-01-02")
	paused, err := s.store.IsDatePaused(ctx, fc.family.ID, date)
	if err != nil {
		return nil, err
	}
	chores, err := s.store.ListChores(ctx, fc.family.ID, 0, false)
	if err != nil {
		return nil, err
	}
	done, err := s.todaysCompletionSet(ctx, fc.family.ID, date)
	if err != nil {
		return nil, err
	}
	chartRows, err := s.store.ListStarCharts(ctx, fc.family.ID, false)
	if err != nil {
		return nil, err
	}
	chartNames := map[int]string{}
	for i := range chartRows {
		chartNames[chartRows[i].ID] = chartRows[i].Name
	}
	defaultChartID, err := s.store.GetDefaultStarChartID(ctx, fc.family.ID)
	if err != nil {
		return nil, err
	}
	type agg struct {
		completed, scheduled int
	}
	aggs := map[int]*agg{}
	for i := range chores {
		if !store.IsScheduledOnDate(chores[i].Chore.WeekdayMask, date) {
			continue
		}
		chartID := chores[i].Chore.StarChartID
		if chartID == 0 {
			chartID = defaultChartID
		}
		for _, assign := range chores[i].Assignments {
			if assign.ChildMemberID != memberID {
				continue
			}
			entry := aggs[chartID]
			if entry == nil {
				entry = &agg{}
				aggs[chartID] = entry
			}
			entry.scheduled++
			if done[fmt.Sprintf("%d:%d", chores[i].Chore.ID, assign.ChildMemberID)] {
				entry.completed++
			}
		}
	}
	out := make([]*apiv1.StarChartDayProgress, 0, len(aggs))
	for chartID, entry := range aggs {
		if entry.scheduled == 0 {
			continue
		}
		name := chartNames[chartID]
		if name == "" {
			name = "Star Chart"
		}
		out = append(out, &apiv1.StarChartDayProgress{
			StarChartId:   int32(chartID),
			StarChartName: name,
			Completed:     int32(entry.completed),
			Scheduled:     int32(entry.scheduled),
			Paused:        paused,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].StarChartName < out[j].StarChartName
	})
	return out, nil
}

func (s *Server) GetMemberTodaysChores(ctx context.Context, req *connect.Request[apiv1.GetMemberTodaysChoresRequest]) (*connect.Response[apiv1.GetMemberTodaysChoresResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	target, err := s.store.GetMemberByID(ctx, int(req.Msg.MemberId))
	if err != nil || target == nil || !isFamilyStarMember(target, fc.family.ID) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("member not found"))
	}
	if !s.canViewMember(fc.au, fc.member, target) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("forbidden"))
	}
	chores, err := s.listTodaysChores(ctx, fc, target.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&apiv1.GetMemberTodaysChoresResponse{TodaysChores: chores}), nil
}

func (s *Server) authorizeChoreToggle(fc *familyContext, childMemberID int) error {
	if fc.au.HasPermission(rbac.PermissionChoresComplete) {
		return nil
	}
	if fc.member != nil && fc.member.ID == childMemberID {
		return nil
	}
	return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("cannot complete chores for another person"))
}
