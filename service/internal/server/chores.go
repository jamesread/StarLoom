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

// toProtoChore maps a stored chore to the API type.
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
		SortOrder:      int32(cw.Chore.SortOrder),
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
	fc, err := s.requireFamilyPermission(ctx, rbac.PermissionChoresViewFamily)
	if err != nil {
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
	fc, err := s.requireFamilyWritePermission(ctx, rbac.PermissionChoresManage)
	if err != nil {
		return nil, err
	}
	return s.createChoreForFamily(ctx, fc, req.Msg)
}

func (s *Server) createChoreForFamily(ctx context.Context, fc *familyContext, msg *apiv1.CreateChoreRequest) (*connect.Response[apiv1.CreateChoreResponse], error) {
	title, reward, err := normalizeChoreInput(msg.Title, int(msg.StarReward))
	if err != nil {
		return nil, err
	}
	childIDs, err := s.validateChoreMemberIDs(ctx, fc.family.ID, msg.ChildMemberIds)
	if err != nil {
		return nil, err
	}
	starChartID, _, err := s.resolveStarChartID(ctx, fc.family.ID, int(msg.StarChartId))
	if err != nil {
		return nil, err
	}
	id, err := s.store.CreateChore(ctx, fc.family.ID, starChartID, title, reward, weekdaysToMask(msg.Weekdays), childIDs)
	if err != nil {
		return nil, mapStoreError(err)
	}
	cw, _ := s.store.GetChoreByID(ctx, id)
	return connect.NewResponse(&apiv1.CreateChoreResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Chore created"},
		Chore:            toProtoChore(cw),
	}), nil
}

func normalizeChoreInput(title string, reward int) (string, int, error) {
	if title == "" {
		return "", 0, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("title required"))
	}
	if reward <= 0 {
		reward = 1
	}
	if reward > cvar.MaxAwardStars {
		return "", 0, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("star_reward must be 1-%d", cvar.MaxAwardStars))
	}
	return title, reward, nil
}

func (s *Server) UpdateChore(ctx context.Context, req *connect.Request[apiv1.UpdateChoreRequest]) (*connect.Response[apiv1.UpdateChoreResponse], error) {
	fc, err := s.requireFamilyWritePermission(ctx, rbac.PermissionChoresManage)
	if err != nil {
		return nil, err
	}
	return s.updateChoreForFamily(ctx, fc, req.Msg)
}

func (s *Server) updateChoreForFamily(ctx context.Context, fc *familyContext, msg *apiv1.UpdateChoreRequest) (*connect.Response[apiv1.UpdateChoreResponse], error) {
	cw, err := s.familyChore(ctx, fc.family.ID, int(msg.Id))
	if err != nil {
		return nil, err
	}
	params, err := s.updateChoreParams(ctx, fc.family.ID, cw, msg)
	if err != nil {
		return nil, err
	}
	if err := s.store.UpdateChore(ctx, cw.Chore.ID, params.starChartID, msg.Title, params.reward, weekdaysToMask(msg.Weekdays), msg.Active, params.childIDs); err != nil {
		return nil, mapStoreError(err)
	}
	updated, _ := s.store.GetChoreByID(ctx, cw.Chore.ID)
	return connect.NewResponse(&apiv1.UpdateChoreResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Chore updated"},
		Chore:            toProtoChore(updated),
	}), nil
}

type choreUpdateParams struct {
	childIDs    []int
	reward      int
	starChartID int
}

func (s *Server) updateChoreParams(ctx context.Context, familyID int, cw *store.ChoreWithAssignments, msg *apiv1.UpdateChoreRequest) (choreUpdateParams, error) {
	reward := int(msg.StarReward)
	if reward <= 0 {
		reward = 1
	}
	childIDs, err := s.validateChoreMemberIDs(ctx, familyID, msg.ChildMemberIds)
	if err != nil {
		return choreUpdateParams{}, err
	}
	starChartID, err := s.choreStarChartID(ctx, familyID, cw.Chore.StarChartID, int(msg.StarChartId))
	if err != nil {
		return choreUpdateParams{}, err
	}
	return choreUpdateParams{reward: reward, childIDs: childIDs, starChartID: starChartID}, nil
}

func (s *Server) familyChore(ctx context.Context, familyID, choreID int) (*store.ChoreWithAssignments, error) {
	cw, err := s.store.GetChoreByID(ctx, choreID)
	if err != nil || cw == nil || cw.Chore.FamilyID != familyID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("chore not found"))
	}
	return cw, nil
}

func (s *Server) choreStarChartID(ctx context.Context, familyID, currentID, requestedID int) (int, error) {
	starChartID := currentID
	if requestedID > 0 {
		id, _, err := s.resolveStarChartID(ctx, familyID, requestedID)
		if err != nil {
			return 0, err
		}
		starChartID = id
	}
	if starChartID != 0 {
		return starChartID, nil
	}
	id, _, err := s.resolveStarChartID(ctx, familyID, 0)
	return id, err
}

// ReorderChores stores a parent's chore display order.
func (s *Server) ReorderChores(ctx context.Context, req *connect.Request[apiv1.ReorderChoresRequest]) (*connect.Response[apiv1.ReorderChoresResponse], error) {
	fc, err := s.requireFamilyWritePermission(ctx, rbac.PermissionChoresManage)
	if err != nil {
		return nil, err
	}
	ids, err := s.validateReorderChoreIDs(ctx, fc.family.ID, req.Msg.ChoreIds)
	if err != nil {
		return nil, err
	}
	if err := s.store.ReorderChores(ctx, fc.family.ID, ids); err != nil {
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.ReorderChoresResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Chore order saved"},
	}), nil
}

func (s *Server) DeleteChore(ctx context.Context, req *connect.Request[apiv1.DeleteChoreRequest]) (*connect.Response[apiv1.DeleteChoreResponse], error) {
	fc, err := s.requireFamilyWritePermission(ctx, rbac.PermissionChoresManage)
	if err != nil {
		return nil, err
	}
	cw, err := s.familyChore(ctx, fc.family.ID, int(req.Msg.Id))
	if err != nil {
		return nil, err
	}
	if err := s.store.DeactivateChore(ctx, cw.Chore.ID); err != nil {
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.DeleteChoreResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Chore deactivated"},
	}), nil
}

func (s *Server) ListChorePauses(ctx context.Context, _ *connect.Request[apiv1.ListChorePausesRequest]) (*connect.Response[apiv1.ListChorePausesResponse], error) {
	fc, err := s.requireFamilyPermission(ctx, rbac.PermissionChoresViewFamily)
	if err != nil {
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
	fc, err := s.requireFamilyWritePermission(ctx, rbac.PermissionChoresManage)
	if err != nil {
		return nil, err
	}
	start, end, err := validatePauseDates(req.Msg.StartDate, req.Msg.EndDate)
	if err != nil {
		return nil, err
	}
	id, err := s.store.CreateChorePause(ctx, fc.family.ID, start, end, req.Msg.Reason)
	if err != nil {
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.CreateChorePauseResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Chores paused"},
		Pause:            toProtoChorePause(s.findChorePause(ctx, fc.family.ID, id)),
	}), nil
}

func validatePauseDates(start, end string) (string, string, error) {
	if start == "" || end == "" {
		return "", "", connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("start_date and end_date required"))
	}
	if start > end {
		return "", "", connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("start_date must be on or before end_date"))
	}
	return start, end, nil
}

func (s *Server) findChorePause(ctx context.Context, familyID, id int) *store.ChorePauseRow {
	pauses, _ := s.store.ListChorePauses(ctx, familyID)
	for i := range pauses {
		if pauses[i].ID == id {
			return &pauses[i]
		}
	}
	return nil
}

func (s *Server) DeleteChorePause(ctx context.Context, req *connect.Request[apiv1.DeleteChorePauseRequest]) (*connect.Response[apiv1.DeleteChorePauseResponse], error) {
	fc, err := s.requireFamilyWritePermission(ctx, rbac.PermissionChoresManage)
	if err != nil {
		return nil, err
	}
	if err := s.requireChorePause(ctx, fc.family.ID, int(req.Msg.Id)); err != nil {
		return nil, err
	}
	if err := s.store.DeleteChorePause(ctx, int(req.Msg.Id)); err != nil {
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.DeleteChorePauseResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Pause removed"},
	}), nil
}

func (s *Server) requireChorePause(ctx context.Context, familyID, pauseID int) error {
	pauses, err := s.store.ListChorePauses(ctx, familyID)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	for _, p := range pauses {
		if p.ID == pauseID {
			return nil
		}
	}
	return connect.NewError(connect.CodeNotFound, fmt.Errorf("pause not found"))
}

func (s *Server) GetWeeklyStarChart(ctx context.Context, req *connect.Request[apiv1.GetWeeklyStarChartRequest]) (*connect.Response[apiv1.GetWeeklyStarChartResponse], error) {
	fc, err := s.requireFamilyPermissionAny(ctx, rbac.PermissionChoresViewFamily, rbac.PermissionStarsViewOwn)
	if err != nil {
		return nil, err
	}
	return s.buildWeeklyStarChart(ctx, fc, req.Msg)
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
	if _, err = s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if err = s.authorizeChoreToggle(fc, int(req.Msg.ChildMemberId)); err != nil {
		return nil, err
	}
	return s.toggleChoreCompletion(ctx, fc, int(req.Msg.ChoreId), int(req.Msg.ChildMemberId), req.Msg.Date, true)
}

func (s *Server) UncompleteChore(ctx context.Context, req *connect.Request[apiv1.UncompleteChoreRequest]) (*connect.Response[apiv1.UncompleteChoreResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err = s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if err = s.authorizeChoreToggle(fc, int(req.Msg.ChildMemberId)); err != nil {
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
	cw, assign, err := s.loadToggleChoreContext(ctx, fc, choreID, childMemberID, date)
	if err != nil {
		return nil, err
	}
	if complete {
		return s.completeChoreAssignment(ctx, fc, cw, assign, childMemberID, date)
	}
	return s.uncompleteChoreAssignment(ctx, fc, cw, assign, childMemberID, date)
}

// validateReorderChoreIDs rejects empty, duplicate and foreign chore ids.
func (s *Server) validateReorderChoreIDs(ctx context.Context, familyID int, ids []int32) ([]int, error) {
	if len(ids) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("at least one chore required"))
	}
	return s.uniqueFamilyChoreIDs(ctx, familyID, ids)
}

func (s *Server) uniqueFamilyChoreIDs(ctx context.Context, familyID int, ids []int32) ([]int, error) {
	out := make([]int, 0, len(ids))
	seen := map[int]bool{}
	for _, id := range ids {
		if err := appendUniqueFamilyChore(ctx, s, familyID, int(id), seen, &out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func appendUniqueFamilyChore(ctx context.Context, s *Server, familyID, id int, seen map[int]bool, out *[]int) error {
	if seen[id] {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("duplicate chore id %d", id))
	}
	cw, err := s.store.GetChoreByID(ctx, id)
	if err != nil || cw == nil || cw.Chore.FamilyID != familyID {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid chore id %d", id))
	}
	seen[id] = true
	*out = append(*out, id)
	return nil
}

func (s *Server) validateChoreMemberIDs(ctx context.Context, familyID int, ids []int32) ([]int, error) {
	if len(ids) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("at least one person required"))
	}
	return s.uniqueFamilyMemberIDs(ctx, familyID, ids)
}

func (s *Server) uniqueFamilyMemberIDs(ctx context.Context, familyID int, ids []int32) ([]int, error) {
	out := make([]int, 0, len(ids))
	seen := map[int]bool{}
	for _, id := range ids {
		if err := appendUniqueFamilyMember(ctx, s, familyID, int(id), seen, &out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func appendUniqueFamilyMember(ctx context.Context, s *Server, familyID, id int, seen map[int]bool, out *[]int) error {
	if seen[id] {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("duplicate member id %d", id))
	}
	member, err := s.store.GetMemberByID(ctx, id)
	if err != nil || !isFamilyStarMember(member, familyID) {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid member id %d", id))
	}
	seen[id] = true
	*out = append(*out, id)
	return nil
}
