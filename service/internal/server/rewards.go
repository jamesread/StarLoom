package server

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/rbac"
	"github.com/jamesread/starapp/service/internal/rewards"
	"github.com/jamesread/starapp/service/internal/store"
)

func (s *Server) ListRewards(ctx context.Context, req *connect.Request[apiv1.ListRewardsRequest]) (*connect.Response[apiv1.ListRewardsResponse], error) {
	fc, err := s.requireRewardsListAccess(ctx)
	if err != nil {
		return nil, err
	}
	includeInactive := req.Msg.IncludeInactive && fc.au.HasPermission(rbac.PermissionRewardsManage)
	rows, err := s.store.ListRewards(ctx, fc.family.ID, includeInactive)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := &apiv1.ListRewardsResponse{}
	for i := range rows {
		out.Rewards = append(out.Rewards, toProtoReward(&rows[i]))
	}
	return connect.NewResponse(out), nil
}

func (s *Server) requireRewardsListAccess(ctx context.Context) (*familyContext, error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if !fc.au.HasPermission(rbac.PermissionRewardsView) && !fc.au.HasPermission(rbac.PermissionRewardsManage) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("forbidden"))
	}
	return fc, nil
}

func (s *Server) CreateReward(ctx context.Context, req *connect.Request[apiv1.CreateRewardRequest]) (*connect.Response[apiv1.CreateRewardResponse], error) {
	fc, err := s.requireFamilyWritePermission(ctx, rbac.PermissionRewardsManage)
	if err != nil {
		return nil, err
	}
	title, expr, err := validateRewardInput(req.Msg.Title, req.Msg.CostStars, req.Msg.AvailabilityExpression)
	if err != nil {
		return nil, err
	}
	approval := req.Msg.ApprovalRequired || s.redemptionApprovalDefault(ctx)
	id, err := s.store.CreateReward(ctx, fc.family.ID, title, req.Msg.Description, int(req.Msg.CostStars), approval, expr, req.Msg.MultiSpend)
	if err != nil {
		return nil, mapStoreError(err)
	}
	reward, _ := s.store.GetRewardByID(ctx, id)
	return connect.NewResponse(&apiv1.CreateRewardResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Reward created"},
		Reward:           toProtoReward(reward),
	}), nil
}

func (s *Server) UpdateReward(ctx context.Context, req *connect.Request[apiv1.UpdateRewardRequest]) (*connect.Response[apiv1.UpdateRewardResponse], error) {
	fc, err := s.requireFamilyWritePermission(ctx, rbac.PermissionRewardsManage)
	if err != nil {
		return nil, err
	}
	reward, err := s.familyReward(ctx, fc.family.ID, int(req.Msg.Id))
	if err != nil {
		return nil, err
	}
	title, expr, err := validateRewardInput(req.Msg.Title, req.Msg.CostStars, req.Msg.AvailabilityExpression)
	if err != nil {
		return nil, err
	}
	if err := s.store.UpdateReward(ctx, reward.ID, title, req.Msg.Description, int(req.Msg.CostStars), req.Msg.Active, req.Msg.ApprovalRequired, expr, req.Msg.MultiSpend); err != nil {
		return nil, mapStoreError(err)
	}
	reward, _ = s.store.GetRewardByID(ctx, reward.ID)
	return connect.NewResponse(&apiv1.UpdateRewardResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Reward updated"},
		Reward:           toProtoReward(reward),
	}), nil
}

func (s *Server) DeleteReward(ctx context.Context, req *connect.Request[apiv1.DeleteRewardRequest]) (*connect.Response[apiv1.DeleteRewardResponse], error) {
	fc, err := s.requireFamilyWritePermission(ctx, rbac.PermissionRewardsManage)
	if err != nil {
		return nil, err
	}
	reward, err := s.familyReward(ctx, fc.family.ID, int(req.Msg.Id))
	if err != nil {
		return nil, err
	}
	if err := s.store.DeactivateReward(ctx, reward.ID); err != nil {
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.DeleteRewardResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Reward deactivated"},
	}), nil
}

func validateRewardInput(title string, costStars int32, availabilityExpression string) (string, string, error) {
	title = strings.TrimSpace(title)
	if title == "" || costStars <= 0 {
		return "", "", connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("title and positive cost required"))
	}
	expr := strings.TrimSpace(availabilityExpression)
	if err := rewards.ValidateAvailabilityExpression(expr); err != nil {
		return "", "", connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid availability expression: %w", err))
	}
	return title, expr, nil
}

func (s *Server) familyReward(ctx context.Context, familyID, rewardID int) (*store.RewardRow, error) {
	reward, err := s.store.GetRewardByID(ctx, rewardID)
	if err != nil || reward == nil || reward.FamilyID != familyID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("reward not found"))
	}
	return reward, nil
}
