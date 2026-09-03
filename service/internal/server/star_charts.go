package server

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/rbac"
	"github.com/jamesread/starapp/service/internal/store"
)

func toProtoStarChart(sc *store.StarChartRow, choreCount int) *apiv1.StarChart {
	if sc == nil {
		return nil
	}
	return &apiv1.StarChart{
		Id:         int32(sc.ID),
		FamilyId:   int32(sc.FamilyID),
		Name:       sc.Name,
		SortOrder:  int32(sc.SortOrder),
		Active:     sc.Active,
		CreatedAt:  sc.CreatedAt,
		ChoreCount: int32(choreCount),
	}
}

func (s *Server) resolveStarChartID(ctx context.Context, familyID int, starChartID int) (int, *store.StarChartRow, error) {
	if starChartID > 0 {
		sc, err := s.store.GetStarChartByID(ctx, starChartID)
		if err != nil {
			return 0, nil, connect.NewError(connect.CodeInternal, err)
		}
		if sc == nil || sc.FamilyID != familyID {
			return 0, nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("star chart not found"))
		}
		return sc.ID, sc, nil
	}
	id, err := s.store.GetDefaultStarChartID(ctx, familyID)
	if err != nil {
		return 0, nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("no star chart found"))
	}
	sc, err := s.store.GetStarChartByID(ctx, id)
	if err != nil {
		return 0, nil, connect.NewError(connect.CodeInternal, err)
	}
	return id, sc, nil
}

func (s *Server) ListStarCharts(ctx context.Context, req *connect.Request[apiv1.ListStarChartsRequest]) (*connect.Response[apiv1.ListStarChartsResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionChoresViewFamily); err != nil {
		if _, err2 := s.requirePermission(ctx, rbac.PermissionStarsViewOwn); err2 != nil {
			return nil, err
		}
	}
	memberFilter := s.chartMemberFilter(fc)
	if req.Msg.GetAssignedToMe() && fc.member != nil {
		memberFilter = fc.member.ID
	}
	rows, err := s.store.ListStarCharts(ctx, fc.family.ID, req.Msg.IncludeInactive)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := &apiv1.ListStarChartsResponse{}
	for i := range rows {
		var count int
		if memberFilter != 0 {
			count, err = s.store.CountChoresForStarChartAndMember(ctx, rows[i].ID, memberFilter)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, err)
			}
			if count == 0 {
				continue
			}
		} else {
			count, err = s.store.CountChoresForStarChart(ctx, rows[i].ID)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, err)
			}
		}
		out.StarCharts = append(out.StarCharts, toProtoStarChart(&rows[i], count))
	}
	return connect.NewResponse(out), nil
}

func (s *Server) CreateStarChart(ctx context.Context, req *connect.Request[apiv1.CreateStarChartRequest]) (*connect.Response[apiv1.CreateStarChartResponse], error) {
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
	name := strings.TrimSpace(req.Msg.Name)
	if name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name required"))
	}
	id, err := s.store.CreateStarChart(ctx, fc.family.ID, name, int(req.Msg.SortOrder))
	if err != nil {
		return nil, mapStoreError(err)
	}
	sc, _ := s.store.GetStarChartByID(ctx, id)
	return connect.NewResponse(&apiv1.CreateStarChartResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Star chart created"},
		StarChart:        toProtoStarChart(sc, 0),
	}), nil
}

func (s *Server) UpdateStarChart(ctx context.Context, req *connect.Request[apiv1.UpdateStarChartRequest]) (*connect.Response[apiv1.UpdateStarChartResponse], error) {
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
	sc, err := s.store.GetStarChartByID(ctx, int(req.Msg.Id))
	if err != nil || sc == nil || sc.FamilyID != fc.family.ID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("star chart not found"))
	}
	name := strings.TrimSpace(req.Msg.Name)
	if name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name required"))
	}
	if err := s.store.UpdateStarChart(ctx, sc.ID, name, int(req.Msg.SortOrder), req.Msg.Active); err != nil {
		return nil, mapStoreError(err)
	}
	updated, _ := s.store.GetStarChartByID(ctx, sc.ID)
	count, _ := s.store.CountChoresForStarChart(ctx, sc.ID)
	return connect.NewResponse(&apiv1.UpdateStarChartResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Star chart updated"},
		StarChart:        toProtoStarChart(updated, count),
	}), nil
}

func (s *Server) DeleteStarChart(ctx context.Context, req *connect.Request[apiv1.DeleteStarChartRequest]) (*connect.Response[apiv1.DeleteStarChartResponse], error) {
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
	sc, err := s.store.GetStarChartByID(ctx, int(req.Msg.Id))
	if err != nil || sc == nil || sc.FamilyID != fc.family.ID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("star chart not found"))
	}
	count, err := s.store.CountChoresForStarChart(ctx, sc.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if count > 0 {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("star chart has assigned chores"))
	}
	charts, err := s.store.ListStarCharts(ctx, fc.family.ID, true)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if len(charts) <= 1 {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("cannot delete the only star chart"))
	}
	if err := s.store.DeleteStarChart(ctx, sc.ID); err != nil {
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.DeleteStarChartResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Star chart deleted"},
	}), nil
}
