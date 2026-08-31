package server

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/cvar"
	"github.com/jamesread/starapp/service/internal/rbac"
	"github.com/jamesread/starapp/service/internal/store"
)

func weekdaysToMask(weekdays []int32) int {
	days := make([]int, len(weekdays))
	for i, d := range weekdays {
		days[i] = int(d)
	}
	return store.WeekdaysToMask(days)
}

func maskToWeekdaysProto(mask int) []int32 {
	days := store.MaskToWeekdays(mask)
	out := make([]int32, len(days))
	for i, d := range days {
		out[i] = int32(d)
	}
	return out
}

func toProtoChore(cw *store.ChoreWithAssignments) *apiv1.Chore {
	if cw == nil {
		return nil
	}
	childIDs := make([]int32, len(cw.Assignments))
	for i, a := range cw.Assignments {
		childIDs[i] = int32(a.ChildMemberID)
	}
	return &apiv1.Chore{
		Id:             int32(cw.Chore.ID),
		FamilyId:       int32(cw.Chore.FamilyID),
		Title:          cw.Chore.Title,
		StarReward:     int32(cw.Chore.StarReward),
		Weekdays:       maskToWeekdaysProto(cw.Chore.WeekdayMask),
		Active:         cw.Chore.Active,
		ChildMemberIds: childIDs,
		CreatedAt:      cw.Chore.CreatedAt,
		StarChartId:    int32(cw.Chore.StarChartID),
	}
}

func toProtoChorePause(p *store.ChorePauseRow) *apiv1.ChorePause {
	if p == nil {
		return nil
	}
	return &apiv1.ChorePause{
		Id:        int32(p.ID),
		FamilyId:  int32(p.FamilyID),
		StartDate: p.StartDate,
		EndDate:   p.EndDate,
		Reason:    p.Reason,
		CreatedAt: p.CreatedAt,
	}
}

func parseWeekStart(raw string) (string, error) {
	if raw == "" {
		now := time.Now()
		wd := int(now.Weekday())
		if wd == 0 {
			wd = 7
		}
		monday := now.AddDate(0, 0, -(wd - 1))
		return monday.Format("2006-01-02"), nil
	}
	if _, err := time.Parse("2006-01-02", raw); err != nil {
		return "", fmt.Errorf("week_start must be YYYY-MM-DD")
	}
	return raw, nil
}

func (s *Server) ListChores(ctx context.Context, req *connect.Request[apiv1.ListChoresRequest]) (*connect.Response[apiv1.ListChoresResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionChoresViewFamily); err != nil {
		return nil, err
	}
	rows, err := s.store.ListChores(ctx, fc.family.ID, int(req.Msg.StarChartId), req.Msg.IncludeInactive)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := &apiv1.ListChoresResponse{}
	for i := range rows {
		out.Chores = append(out.Chores, toProtoChore(&rows[i]))
	}
	return connect.NewResponse(out), nil
}

func (s *Server) CreateChore(ctx context.Context, req *connect.Request[apiv1.CreateChoreRequest]) (*connect.Response[apiv1.CreateChoreResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionChoresManage); err != nil {
		return nil, err
	}
	title := req.Msg.Title
	if title == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("title required"))
	}
	reward := int(req.Msg.StarReward)
	if reward <= 0 {
		reward = 1
	}
	if reward > cvar.MaxAwardStars {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("star_reward must be 1-%d", cvar.MaxAwardStars))
	}
	childIDs, err := s.validateChoreMemberIDs(ctx, fc.family.ID, req.Msg.ChildMemberIds)
	if err != nil {
		return nil, err
	}
	mask := weekdaysToMask(req.Msg.Weekdays)
	starChartID, _, err := s.resolveStarChartID(ctx, fc.family.ID, int(req.Msg.StarChartId))
	if err != nil {
		return nil, err
	}
	id, err := s.store.CreateChore(ctx, fc.family.ID, starChartID, title, reward, mask, childIDs)
	if err != nil {
		return nil, mapStoreError(err)
	}
	cw, _ := s.store.GetChoreByID(ctx, id)
	return connect.NewResponse(&apiv1.CreateChoreResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Chore created"},
		Chore:            toProtoChore(cw),
	}), nil
}

func (s *Server) UpdateChore(ctx context.Context, req *connect.Request[apiv1.UpdateChoreRequest]) (*connect.Response[apiv1.UpdateChoreResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionChoresManage); err != nil {
		return nil, err
	}
	cw, err := s.store.GetChoreByID(ctx, int(req.Msg.Id))
	if err != nil || cw == nil || cw.Chore.FamilyID != fc.family.ID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("chore not found"))
	}
	reward := int(req.Msg.StarReward)
	if reward <= 0 {
		reward = 1
	}
	childIDs, err := s.validateChoreMemberIDs(ctx, fc.family.ID, req.Msg.ChildMemberIds)
	if err != nil {
		return nil, err
	}
	mask := weekdaysToMask(req.Msg.Weekdays)
	starChartID := cw.Chore.StarChartID
	if req.Msg.StarChartId > 0 {
		id, _, err := s.resolveStarChartID(ctx, fc.family.ID, int(req.Msg.StarChartId))
		if err != nil {
			return nil, err
		}
		starChartID = id
	}
	if starChartID == 0 {
		id, _, err := s.resolveStarChartID(ctx, fc.family.ID, 0)
		if err != nil {
			return nil, err
		}
		starChartID = id
	}
	if err := s.store.UpdateChore(ctx, cw.Chore.ID, starChartID, req.Msg.Title, reward, mask, req.Msg.Active, childIDs); err != nil {
		return nil, mapStoreError(err)
	}
	updated, _ := s.store.GetChoreByID(ctx, cw.Chore.ID)
	return connect.NewResponse(&apiv1.UpdateChoreResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Chore updated"},
		Chore:            toProtoChore(updated),
	}), nil
}

func (s *Server) DeleteChore(ctx context.Context, req *connect.Request[apiv1.DeleteChoreRequest]) (*connect.Response[apiv1.DeleteChoreResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionChoresManage); err != nil {
		return nil, err
	}
	cw, err := s.store.GetChoreByID(ctx, int(req.Msg.Id))
	if err != nil || cw == nil || cw.Chore.FamilyID != fc.family.ID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("chore not found"))
	}
	if err := s.store.DeactivateChore(ctx, cw.Chore.ID); err != nil {
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.DeleteChoreResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Chore deactivated"},
	}), nil
}

func (s *Server) ListChorePauses(ctx context.Context, _ *connect.Request[apiv1.ListChorePausesRequest]) (*connect.Response[apiv1.ListChorePausesResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionChoresViewFamily); err != nil {
		return nil, err
	}
	rows, err := s.store.ListChorePauses(ctx, fc.family.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := &apiv1.ListChorePausesResponse{}
	for i := range rows {
		out.Pauses = append(out.Pauses, toProtoChorePause(&rows[i]))
	}
	return connect.NewResponse(out), nil
}

func (s *Server) CreateChorePause(ctx context.Context, req *connect.Request[apiv1.CreateChorePauseRequest]) (*connect.Response[apiv1.CreateChorePauseResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionChoresManage); err != nil {
		return nil, err
	}
	start, end := req.Msg.StartDate, req.Msg.EndDate
	if start == "" || end == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("start_date and end_date required"))
	}
	if start > end {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("start_date must be on or before end_date"))
	}
	id, err := s.store.CreateChorePause(ctx, fc.family.ID, start, end, req.Msg.Reason)
	if err != nil {
		return nil, mapStoreError(err)
	}
	pauses, _ := s.store.ListChorePauses(ctx, fc.family.ID)
	var pause *store.ChorePauseRow
	for i := range pauses {
		if pauses[i].ID == id {
			pause = &pauses[i]
			break
		}
	}
	return connect.NewResponse(&apiv1.CreateChorePauseResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Chores paused"},
		Pause:            toProtoChorePause(pause),
	}), nil
}

func (s *Server) DeleteChorePause(ctx context.Context, req *connect.Request[apiv1.DeleteChorePauseRequest]) (*connect.Response[apiv1.DeleteChorePauseResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionChoresManage); err != nil {
		return nil, err
	}
	pauses, err := s.store.ListChorePauses(ctx, fc.family.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	found := false
	for _, p := range pauses {
		if p.ID == int(req.Msg.Id) {
			found = true
			break
		}
	}
	if !found {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("pause not found"))
	}
	if err := s.store.DeleteChorePause(ctx, int(req.Msg.Id)); err != nil {
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.DeleteChorePauseResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Pause removed"},
	}), nil
}

func (s *Server) GetWeeklyStarChart(ctx context.Context, req *connect.Request[apiv1.GetWeeklyStarChartRequest]) (*connect.Response[apiv1.GetWeeklyStarChartResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionChoresViewFamily); err != nil {
		if _, err2 := s.requirePermission(ctx, rbac.PermissionStarsViewOwn); err2 != nil {
			return nil, err
		}
	}
	weekStart, err := parseWeekStart(req.Msg.WeekStart)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	dates, err := store.WeekDates(weekStart)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	weekEnd := dates[6]

	starChartID, starChart, err := s.resolveStarChartID(ctx, fc.family.ID, int(req.Msg.StarChartId))
	if err != nil {
		return nil, err
	}

	chores, err := s.store.ListChores(ctx, fc.family.ID, starChartID, false)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	completions, err := s.store.ListCompletionsForWeek(ctx, fc.family.ID, weekStart, weekEnd)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	completionMap := map[string]store.WeeklyChartCompletion{}
	for _, c := range completions {
		key := fmt.Sprintf("%d:%d:%s", c.ChoreID, c.ChildMemberID, c.CompletionDate)
		completionMap[key] = c
	}

	members, _ := s.store.ListMembersByFamily(ctx, fc.family.ID)
	memberByID := map[int]store.FamilyMemberRow{}
	for _, m := range members {
		memberByID[m.ID] = m
	}

	childFilter := s.chartChildFilter(fc)
	out := &apiv1.GetWeeklyStarChartResponse{
		WeekStart:     weekStart,
		WeekEnd:       weekEnd,
		StarChartId:   int32(starChartID),
		StarChartName: starChart.Name,
	}

	for _, cw := range chores {
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
			}
			for _, date := range dates {
				paused, _ := s.store.IsDatePaused(ctx, fc.family.ID, date)
				scheduled := store.IsScheduledOnDate(cw.Chore.WeekdayMask, date) && !paused
				key := fmt.Sprintf("%d:%d:%s", cw.Chore.ID, assign.ChildMemberID, date)
				comp := completionMap[key]
				day := &apiv1.WeeklyStarChartDay{
					Date: date, Scheduled: scheduled, Paused: paused,
					Completed: comp.CompletionDate != "", StarsEarned: int32(comp.StarsEarned),
				}
				childRow.Days = append(childRow.Days, day)
			}
			row.Children = append(row.Children, childRow)
		}
		if len(row.Children) > 0 {
			out.Rows = append(out.Rows, row)
		}
	}

	choreLedgerIDs, _ := s.store.ListChoreLedgerEntryIDs(ctx, fc.family.ID)
	bonusMap, err := s.store.ListBonusStarsForWeek(ctx, fc.family.ID, weekStart, weekEnd, choreLedgerIDs)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	for _, m := range members {
		if !isFamilyStarMember(&m, fc.family.ID) {
			continue
		}
		if childFilter != 0 && m.ID != childFilter {
			continue
		}
		childBonus := bonusMap[m.ID]
		childRow := &apiv1.WeeklyStarChartBonusChild{
			Child: toProtoMember(&m),
		}
		for _, date := range dates {
			childRow.Days = append(childRow.Days, &apiv1.WeeklyStarChartDay{
				Date: date, StarsEarned: int32(childBonus[date]),
			})
		}
		out.BonusChildren = append(out.BonusChildren, childRow)
	}
	return connect.NewResponse(out), nil
}

func (s *Server) chartMemberFilter(fc *familyContext) int {
	if fc.au.HasPermission(rbac.PermissionStarsViewFamily) {
		return 0
	}
	if fc.member != nil && fc.member.Role == store.MemberRoleChild {
		return fc.member.ID
	}
	return 0
}

func (s *Server) chartChildFilter(fc *familyContext) int {
	return s.chartMemberFilter(fc)
}

func (s *Server) CompleteChore(ctx context.Context, req *connect.Request[apiv1.CompleteChoreRequest]) (*connect.Response[apiv1.CompleteChoreResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionChoresComplete); err != nil {
		return nil, err
	}
	return s.toggleChoreCompletion(ctx, fc, int(req.Msg.ChoreId), int(req.Msg.ChildMemberId), req.Msg.Date, true)
}

func (s *Server) UncompleteChore(ctx context.Context, req *connect.Request[apiv1.UncompleteChoreRequest]) (*connect.Response[apiv1.UncompleteChoreResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionChoresComplete); err != nil {
		return nil, err
	}
	resp, err := s.toggleChoreCompletion(ctx, fc, int(req.Msg.ChoreId), int(req.Msg.ChildMemberId), req.Msg.Date, false)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&apiv1.UncompleteChoreResponse{
		StandardResponse: resp.Msg.StandardResponse,
		NewBalance:       resp.Msg.NewBalance,
	}), nil
}

func (s *Server) toggleChoreCompletion(ctx context.Context, fc *familyContext, choreID, childMemberID int, date string, complete bool) (*connect.Response[apiv1.CompleteChoreResponse], error) {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	cw, err := s.store.GetChoreByID(ctx, choreID)
	if err != nil || cw == nil || cw.Chore.FamilyID != fc.family.ID || !cw.Chore.Active {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("chore not found"))
	}
	member, err := s.store.GetMemberByID(ctx, childMemberID)
	if err != nil || !isFamilyStarMember(member, fc.family.ID) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("member not found"))
	}
	assign, err := s.store.GetAssignment(ctx, choreID, childMemberID)
	if err != nil || assign == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("member not assigned to chore"))
	}
	paused, _ := s.store.IsDatePaused(ctx, fc.family.ID, date)
	if paused {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("chores are paused on this date"))
	}
	if !store.IsScheduledOnDate(cw.Chore.WeekdayMask, date) {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("chore not scheduled on this date"))
	}
	existing, err := s.store.GetCompletion(ctx, assign.ID, date)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if complete {
		if existing != nil {
			balance, _ := s.store.GetMemberBalance(ctx, childMemberID)
			return connect.NewResponse(&apiv1.CompleteChoreResponse{
				StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Already completed"},
				NewBalance:       int32(balance),
			}), nil
		}
		createdBy := fc.member.ID
		entryID, err := s.store.InsertLedgerEntry(ctx, store.StarLedgerRow{
			FamilyID: fc.family.ID, ChildMemberID: childMemberID, Amount: cw.Chore.StarReward,
			EntryType: store.LedgerTypeAward, Note: fmt.Sprintf("Chore: %s", cw.Chore.Title),
			CreatedByMemberID: &createdBy,
		})
		if err != nil {
			return nil, mapStoreError(err)
		}
		if _, err := s.store.InsertChoreCompletion(ctx, assign.ID, date, entryID); err != nil {
			return nil, mapStoreError(err)
		}
		s.webhooks.Dispatch(ctx, "stars.awarded", map[string]any{
			"family_id": fc.family.ID, "child_member_id": childMemberID,
			"amount": cw.Chore.StarReward, "note": fmt.Sprintf("Chore: %s", cw.Chore.Title),
			"created_by_member_id": createdBy, "chore_id": choreID,
		})
	} else {
		if existing == nil {
			balance, _ := s.store.GetMemberBalance(ctx, childMemberID)
			return connect.NewResponse(&apiv1.CompleteChoreResponse{
				StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Not completed"},
				NewBalance:       int32(balance),
			}), nil
		}
		createdBy := fc.member.ID
		neg := -cw.Chore.StarReward
		if _, err := s.store.InsertLedgerEntry(ctx, store.StarLedgerRow{
			FamilyID: fc.family.ID, ChildMemberID: childMemberID, Amount: neg,
			EntryType: store.LedgerTypeRevoke, Note: fmt.Sprintf("Undo chore: %s", cw.Chore.Title),
			CreatedByMemberID: &createdBy,
		}); err != nil {
			return nil, mapStoreError(err)
		}
		if err := s.store.DeleteChoreCompletion(ctx, existing.ID); err != nil {
			return nil, mapStoreError(err)
		}
	}
	balance, _ := s.store.GetMemberBalance(ctx, childMemberID)
	msg := "Chore completed"
	if !complete {
		msg = "Chore uncompleted"
	}
	return connect.NewResponse(&apiv1.CompleteChoreResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: msg},
		NewBalance:       int32(balance),
	}), nil
}

func (s *Server) validateChoreMemberIDs(ctx context.Context, familyID int, ids []int32) ([]int, error) {
	if len(ids) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("at least one person required"))
	}
	out := make([]int, 0, len(ids))
	seen := map[int]bool{}
	for _, id := range ids {
		if seen[int(id)] {
			continue
		}
		m, err := s.store.GetMemberByID(ctx, int(id))
		if err != nil || !isFamilyStarMember(m, familyID) {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid member id %d", id))
		}
		seen[int(id)] = true
		out = append(out, int(id))
	}
	return out, nil
}
