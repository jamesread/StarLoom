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
	data, err := s.loadTodaysChoreData(ctx, fc.family.ID, date)
	if err != nil {
		return nil, err
	}
	out := make([]*apiv1.TodaysChore, 0)
	for i := range data.chores {
		out = appendTodaysAssignments(&data.chores[i], memberFilter, date, data.paused, data.done, data.memberByID, data.chartNames, data.defaultChartID, out)
	}
	sortTodaysChores(out)
	return out, nil
}

type todaysChoreData struct {
	done           map[string]bool
	memberByID     map[int]store.FamilyMemberRow
	chartNames     map[int]string
	chores         []store.ChoreWithAssignments
	defaultChartID int
	paused         bool
}

func (s *Server) loadTodaysChoreData(ctx context.Context, familyID int, date string) (*todaysChoreData, error) {
	paused, err := s.store.IsDatePaused(ctx, familyID, date)
	if err != nil {
		return nil, err
	}
	chores, err := s.store.ListChores(ctx, familyID, 0, false)
	if err != nil {
		return nil, err
	}
	done, err := s.todaysCompletionSet(ctx, familyID, date)
	if err != nil {
		return nil, err
	}
	memberByID, chartNames, defaultChartID, err := s.todaysLookups(ctx, familyID)
	if err != nil {
		return nil, err
	}
	return &todaysChoreData{
		paused: paused, chores: chores, done: done,
		memberByID: memberByID, chartNames: chartNames, defaultChartID: defaultChartID,
	}, nil
}

func (s *Server) todaysLookups(ctx context.Context, familyID int) (map[int]store.FamilyMemberRow, map[int]string, int, error) {
	memberByID, err := s.memberByIDMap(ctx, familyID)
	if err != nil {
		return nil, nil, 0, err
	}
	chartNames, err := s.chartNamesMap(ctx, familyID)
	if err != nil {
		return nil, nil, 0, err
	}
	defaultChartID, err := s.store.GetDefaultStarChartID(ctx, familyID)
	if err != nil {
		return nil, nil, 0, err
	}
	return memberByID, chartNames, defaultChartID, nil
}

func (s *Server) memberByIDMap(ctx context.Context, familyID int) (map[int]store.FamilyMemberRow, error) {
	members, err := s.store.ListMembersByFamily(ctx, familyID)
	if err != nil {
		return nil, err
	}
	memberByID := map[int]store.FamilyMemberRow{}
	for _, m := range members {
		memberByID[m.ID] = m
	}
	return memberByID, nil
}

func (s *Server) chartNamesMap(ctx context.Context, familyID int) (map[int]string, error) {
	chartRows, err := s.store.ListStarCharts(ctx, familyID, false)
	if err != nil {
		return nil, err
	}
	chartNames := map[int]string{}
	for i := range chartRows {
		chartNames[chartRows[i].ID] = chartRows[i].Name
	}
	return chartNames, nil
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
	chartID, chartName := resolveChartMeta(cw.Chore.StarChartID, defaultChartID, chartNames)
	for _, assign := range cw.Assignments {
		out = maybeAppendTodaysAssignment(cw, assign, memberFilter, date, paused, done, memberByID, chartID, chartName, out)
	}
	return out
}

func maybeAppendTodaysAssignment(
	cw *store.ChoreWithAssignments,
	assign store.ChoreAssignmentRow,
	memberFilter int,
	date string,
	paused bool,
	done map[string]bool,
	memberByID map[int]store.FamilyMemberRow,
	chartID int,
	chartName string,
	out []*apiv1.TodaysChore,
) []*apiv1.TodaysChore {
	if memberFilter != 0 && assign.ChildMemberID != memberFilter {
		return out
	}
	member := memberByID[assign.ChildMemberID]
	return append(out, &apiv1.TodaysChore{
		ChoreId: int32(cw.Chore.ID), Title: cw.Chore.Title, StarReward: int32(cw.Chore.StarReward),
		ChildMemberId: int32(assign.ChildMemberID), Child: toProtoMember(&member),
		Completed: done[fmt.Sprintf("%d:%d", cw.Chore.ID, assign.ChildMemberID)],
		Paused:    paused, Date: date, StarChartId: int32(chartID), StarChartName: chartName,
	})
}

func resolveChartMeta(chartID, defaultChartID int, chartNames map[int]string) (int, string) {
	if chartID == 0 {
		chartID = defaultChartID
	}
	name := chartNames[chartID]
	if name == "" {
		name = "Star Chart"
	}
	return chartID, name
}

func sortTodaysChores(rows []*apiv1.TodaysChore) {
	sort.SliceStable(rows, func(i, j int) bool {
		leftChart := rows[i].GetStarChartName()
		rightChart := rows[j].GetStarChartName()
		if leftChart != rightChart {
			return leftChart < rightChart
		}
		return rows[i].GetChild().GetDisplayName() < rows[j].GetChild().GetDisplayName()
	})
}

func (s *Server) memberTodayStarChartProgress(ctx context.Context, fc *familyContext, memberID int) ([]*apiv1.StarChartDayProgress, error) {
	date := time.Now().Format("2006-01-02")
	data, err := s.loadTodaysChoreData(ctx, fc.family.ID, date)
	if err != nil {
		return nil, err
	}
	aggs := aggregateMemberChartProgress(data.chores, memberID, date, data.done, data.defaultChartID)
	return progressFromAggs(aggs, data.chartNames, data.paused), nil
}

type chartAgg struct {
	completed, scheduled int
}

func aggregateMemberChartProgress(
	chores []store.ChoreWithAssignments,
	memberID int,
	date string,
	done map[string]bool,
	defaultChartID int,
) map[int]*chartAgg {
	aggs := map[int]*chartAgg{}
	for i := range chores {
		if !store.IsScheduledOnDate(chores[i].Chore.WeekdayMask, date) {
			continue
		}
		chartID := chores[i].Chore.StarChartID
		if chartID == 0 {
			chartID = defaultChartID
		}
		accumulateAssignmentProgress(aggs, &chores[i], chartID, memberID, done)
	}
	return aggs
}

func accumulateAssignmentProgress(aggs map[int]*chartAgg, cw *store.ChoreWithAssignments, chartID, memberID int, done map[string]bool) {
	for _, assign := range cw.Assignments {
		if assign.ChildMemberID != memberID {
			continue
		}
		entry := aggs[chartID]
		if entry == nil {
			entry = &chartAgg{}
			aggs[chartID] = entry
		}
		entry.scheduled++
		if done[fmt.Sprintf("%d:%d", cw.Chore.ID, assign.ChildMemberID)] {
			entry.completed++
		}
	}
}

func progressFromAggs(aggs map[int]*chartAgg, chartNames map[int]string, paused bool) []*apiv1.StarChartDayProgress {
	out := make([]*apiv1.StarChartDayProgress, 0, len(aggs))
	for chartID, entry := range aggs {
		if entry.scheduled == 0 {
			continue
		}
		_, name := resolveChartMeta(chartID, chartID, chartNames)
		out = append(out, &apiv1.StarChartDayProgress{
			StarChartId: int32(chartID), StarChartName: name,
			Completed: int32(entry.completed), Scheduled: int32(entry.scheduled), Paused: paused,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].StarChartName < out[j].StarChartName
	})
	return out
}

func (s *Server) GetMemberTodaysChores(ctx context.Context, req *connect.Request[apiv1.GetMemberTodaysChoresRequest]) (*connect.Response[apiv1.GetMemberTodaysChoresResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	target, err := s.viewableStarMember(ctx, fc, int(req.Msg.MemberId))
	if err != nil {
		return nil, err
	}
	chores, err := s.listTodaysChores(ctx, fc, target.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&apiv1.GetMemberTodaysChoresResponse{TodaysChores: chores}), nil
}

func (s *Server) viewableStarMember(ctx context.Context, fc *familyContext, memberID int) (*store.FamilyMemberRow, error) {
	target, err := s.store.GetMemberByID(ctx, memberID)
	if err != nil || target == nil || !isFamilyStarMember(target, fc.family.ID) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("member not found"))
	}
	if !s.canViewMember(fc.au, fc.member, target) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("forbidden"))
	}
	return target, nil
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
