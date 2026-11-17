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
		Id: int32(sc.ID), FamilyId: int32(sc.FamilyID), Name: sc.Name,
		SortOrder: int32(sc.SortOrder), Active: sc.Active, CreatedAt: sc.CreatedAt,
		ChoreCount: int32(choreCount),
	}
}

func (s *Server) resolveStarChartID(ctx context.Context, familyID int, starChartID int) (int, *store.StarChartRow, error) {
	if starChartID > 0 {
		return s.familyStarChart(ctx, familyID, starChartID)
	}
	return s.defaultStarChart(ctx, familyID)
}

func (s *Server) familyStarChart(ctx context.Context, familyID, starChartID int) (int, *store.StarChartRow, error) {
	sc, err := s.store.GetStarChartByID(ctx, starChartID)
	if err != nil {
		return 0, nil, connect.NewError(connect.CodeInternal, err)
	}
	if sc == nil || sc.FamilyID != familyID {
		return 0, nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("star chart not found"))
	}
	return sc.ID, sc, nil
}

func (s *Server) defaultStarChart(ctx context.Context, familyID int) (int, *store.StarChartRow, error) {
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
	fc, err := s.requireFamilyPermissionAny(ctx, rbac.PermissionChoresViewFamily, rbac.PermissionStarsViewOwn)
	if err != nil {
		return nil, err
	}
	memberFilter := s.listStarChartsMemberFilter(fc, req.Msg.GetAssignedToMe())
	rows, err := s.store.ListStarCharts(ctx, fc.family.ID, req.Msg.IncludeInactive)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return s.buildStarChartsResponse(ctx, rows, memberFilter)
}

func (s *Server) buildStarChartsResponse(ctx context.Context, rows []store.StarChartRow, memberFilter int) (*connect.Response[apiv1.ListStarChartsResponse], error) {
	out := &apiv1.ListStarChartsResponse{}
	for i := range rows {
		count, include, err := s.starChartListCount(ctx, rows[i].ID, memberFilter)
		if err != nil {
			return nil, err
		}
		if include {
			out.StarCharts = append(out.StarCharts, toProtoStarChart(&rows[i], count))
		}
	}
	return connect.NewResponse(out), nil
}

func (s *Server) listStarChartsMemberFilter(fc *familyContext, assignedToMe bool) int {
	memberFilter := s.chartMemberFilter(fc)
	if assignedToMe && fc.member != nil {
		return fc.member.ID
	}
	return memberFilter
}

func (s *Server) starChartListCount(ctx context.Context, chartID, memberFilter int) (int, bool, error) {
	if memberFilter == 0 {
		count, err := s.store.CountChoresForStarChart(ctx, chartID)
		return count, true, wrapInternal(err)
	}
	count, err := s.store.CountChoresForStarChartAndMember(ctx, chartID, memberFilter)
	if err != nil {
		return 0, false, connect.NewError(connect.CodeInternal, err)
	}
	return count, count > 0, nil
}

func wrapInternal(err error) error {
	if err == nil {
		return nil
	}
	return connect.NewError(connect.CodeInternal, err)
}

func (s *Server) CreateStarChart(ctx context.Context, req *connect.Request[apiv1.CreateStarChartRequest]) (*connect.Response[apiv1.CreateStarChartResponse], error) {
	fc, err := s.requireFamilyWritePermission(ctx, rbac.PermissionChoresManage)
	if err != nil {
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
	fc, err := s.requireFamilyWritePermission(ctx, rbac.PermissionChoresManage)
	if err != nil {
		return nil, err
	}
	sc, err := s.requireFamilyStarChart(ctx, fc.family.ID, int(req.Msg.Id))
	if err != nil {
		return nil, err
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
	fc, err := s.requireFamilyWritePermission(ctx, rbac.PermissionChoresManage)
	if err != nil {
		return nil, err
	}
	sc, err := s.requireFamilyStarChart(ctx, fc.family.ID, int(req.Msg.Id))
	if err != nil {
		return nil, err
	}
	if err := s.ensureStarChartDeletable(ctx, fc.family.ID, sc.ID); err != nil {
		return nil, err
	}
	if err := s.store.DeleteStarChart(ctx, sc.ID); err != nil {
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.DeleteStarChartResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Star chart deleted"},
	}), nil
}

func (s *Server) requireFamilyStarChart(ctx context.Context, familyID, id int) (*store.StarChartRow, error) {
	sc, err := s.store.GetStarChartByID(ctx, id)
	if err != nil || sc == nil || sc.FamilyID != familyID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("star chart not found"))
	}
	return sc, nil
}

func (s *Server) ensureStarChartDeletable(ctx context.Context, familyID, chartID int) error {
	count, err := s.store.CountChoresForStarChart(ctx, chartID)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if count > 0 {
		return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("star chart has assigned chores"))
	}
	charts, err := s.store.ListStarCharts(ctx, familyID, true)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if len(charts) <= 1 {
		return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("cannot delete the only star chart"))
	}
	return nil
}
