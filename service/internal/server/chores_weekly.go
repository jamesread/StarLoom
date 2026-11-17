package server

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/store"
)

func (s *Server) buildWeeklyStarChart(
	ctx context.Context,
	fc *familyContext,
	msg *apiv1.GetWeeklyStarChartRequest,
) (*connect.Response[apiv1.GetWeeklyStarChartResponse], error) {
	weekStart, dates, err := weeklyChartDates(msg.WeekStart)
	if err != nil {
		return nil, err
	}
	starChartID, starChart, err := s.resolveStarChartID(ctx, fc.family.ID, int(msg.StarChartId))
	if err != nil {
		return nil, err
	}
	chores, completionMap, memberByID, members, err := s.loadWeeklyChartData(ctx, fc.family.ID, starChartID, weekStart, dates[6])
	if err != nil {
		return nil, err
	}
	childFilter := s.chartChildFilter(fc)
	out := &apiv1.GetWeeklyStarChartResponse{
		WeekStart:     weekStart,
		WeekEnd:       dates[6],
		StarChartId:   int32(starChartID),
		StarChartName: starChart.Name,
	}
	out.Rows = buildWeeklyChartRows(ctx, s, fc.family.ID, chores, dates, completionMap, memberByID, childFilter)
	bonusChildren, err := s.buildWeeklyBonusChildren(ctx, fc.family.ID, members, dates, weekStart, dates[6], childFilter)
	if err != nil {
		return nil, err
	}
	out.BonusChildren = bonusChildren
	return connect.NewResponse(out), nil
}

func weeklyChartDates(rawWeekStart string) (weekStart string, dates []string, err error) {
	weekStart, err = parseWeekStart(rawWeekStart)
	if err != nil {
		return "", nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	dates, err = store.WeekDates(weekStart)
	if err != nil {
		return "", nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return weekStart, dates, nil
}

func (s *Server) loadWeeklyChartData(
	ctx context.Context,
	familyID, starChartID int,
	weekStart, weekEnd string,
) (
	chores []store.ChoreWithAssignments,
	completionMap map[string]store.WeeklyChartCompletion,
	memberByID map[int]store.FamilyMemberRow,
	members []store.FamilyMemberRow,
	err error,
) {
	chores, err = s.store.ListChores(ctx, familyID, starChartID, false)
	if err != nil {
		return nil, nil, nil, nil, connect.NewError(connect.CodeInternal, err)
	}
	completions, err := s.store.ListCompletionsForWeek(ctx, familyID, weekStart, weekEnd)
	if err != nil {
		return nil, nil, nil, nil, connect.NewError(connect.CodeInternal, err)
	}
	completionMap = map[string]store.WeeklyChartCompletion{}
	for _, c := range completions {
		key := fmt.Sprintf("%d:%d:%s", c.ChoreID, c.ChildMemberID, c.CompletionDate)
		completionMap[key] = c
	}
	members, _ = s.store.ListMembersByFamily(ctx, familyID)
	memberByID = map[int]store.FamilyMemberRow{}
	for _, m := range members {
		memberByID[m.ID] = m
	}
	return chores, completionMap, memberByID, members, nil
}

func buildWeeklyChartRows(
	ctx context.Context,
	s *Server,
	familyID int,
	chores []store.ChoreWithAssignments,
	dates []string,
	completionMap map[string]store.WeeklyChartCompletion,
	memberByID map[int]store.FamilyMemberRow,
	childFilter int,
) []*apiv1.WeeklyStarChartRow {
	rows := make([]*apiv1.WeeklyStarChartRow, 0, len(chores))
	for _, cw := range chores {
		row := weeklyRowForChore(ctx, s, familyID, cw, dates, completionMap, memberByID, childFilter)
		if len(row.Children) > 0 {
			rows = append(rows, row)
		}
	}
	return rows
}

func weeklyRowForChore(
	ctx context.Context,
	s *Server,
	familyID int,
	cw store.ChoreWithAssignments,
	dates []string,
	completionMap map[string]store.WeeklyChartCompletion,
	memberByID map[int]store.FamilyMemberRow,
	childFilter int,
) *apiv1.WeeklyStarChartRow {
	row := &apiv1.WeeklyStarChartRow{
		ChoreId:    int32(cw.Chore.ID),
		Title:      cw.Chore.Title,
		StarReward: int32(cw.Chore.StarReward),
		Weekdays:   maskToWeekdaysProto(cw.Chore.WeekdayMask),
	}
	for _, assign := range cw.Assignments {
		if childFilter != 0 && assign.ChildMemberID != childFilter {
			continue
		}
		child := memberByID[assign.ChildMemberID]
		childRow := &apiv1.WeeklyStarChartChild{
			AssignmentId: int32(assign.ID),
			Child:        toProtoMember(&child),
			Days:         weeklyDaysForAssignment(ctx, s, familyID, cw, assign.ChildMemberID, dates, completionMap),
		}
		row.Children = append(row.Children, childRow)
	}
	return row
}

func weeklyDaysForAssignment(
	ctx context.Context,
	s *Server,
	familyID int,
	cw store.ChoreWithAssignments,
	childMemberID int,
	dates []string,
	completionMap map[string]store.WeeklyChartCompletion,
) []*apiv1.WeeklyStarChartDay {
	days := make([]*apiv1.WeeklyStarChartDay, 0, len(dates))
	for _, date := range dates {
		paused, _ := s.store.IsDatePaused(ctx, familyID, date)
		scheduled := store.IsScheduledOnDate(cw.Chore.WeekdayMask, date) && !paused
		key := fmt.Sprintf("%d:%d:%s", cw.Chore.ID, childMemberID, date)
		comp := completionMap[key]
		days = append(days, &apiv1.WeeklyStarChartDay{
			Date: date, Scheduled: scheduled, Paused: paused,
			Completed: comp.CompletionDate != "", StarsEarned: int32(comp.StarsEarned),
		})
	}
	return days
}

func (s *Server) buildWeeklyBonusChildren(
	ctx context.Context,
	familyID int,
	members []store.FamilyMemberRow,
	dates []string,
	weekStart, weekEnd string,
	childFilter int,
) ([]*apiv1.WeeklyStarChartBonusChild, error) {
	choreLedgerIDs, _ := s.store.ListChoreLedgerEntryIDs(ctx, familyID)
	bonusMap, err := s.store.ListBonusStarsForWeek(ctx, familyID, weekStart, weekEnd, choreLedgerIDs)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := make([]*apiv1.WeeklyStarChartBonusChild, 0)
	for _, m := range members {
		if !includeBonusMember(&m, familyID, childFilter) {
			continue
		}
		childBonus := bonusMap[m.ID]
		childRow := &apiv1.WeeklyStarChartBonusChild{Child: toProtoMember(&m)}
		for _, date := range dates {
			childRow.Days = append(childRow.Days, &apiv1.WeeklyStarChartDay{
				Date: date, StarsEarned: int32(childBonus[date]),
			})
		}
		out = append(out, childRow)
	}
	return out, nil
}

func includeBonusMember(m *store.FamilyMemberRow, familyID, childFilter int) bool {
	if !isFamilyStarMember(m, familyID) {
		return false
	}
	return childFilter == 0 || m.ID == childFilter
}
