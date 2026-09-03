package server

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/rbac"
	"github.com/jamesread/starapp/service/internal/store"
)

func (s *Server) RequestRedemption(ctx context.Context, req *connect.Request[apiv1.RequestRedemptionRequest]) (*connect.Response[apiv1.RequestRedemptionResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionRedemptionsRequest); err != nil {
		return nil, err
	}
	childID := int(req.Msg.ChildMemberId)
	if childID == 0 {
		childID = fc.member.ID
	}
	child, err := s.store.GetMemberByID(ctx, childID)
	if err != nil || !isFamilyStarMember(child, fc.family.ID) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("member not found"))
	}
	if !fc.au.HasPermission(rbac.PermissionStarsViewFamily) && child.ID != fc.member.ID {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("forbidden"))
	}
	reward, err := s.store.GetRewardByID(ctx, int(req.Msg.RewardId))
	if err != nil || reward == nil || reward.FamilyID != fc.family.ID || !reward.Active {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("reward not found"))
	}
	if reward.ApprovalRequired && child.ID == fc.member.ID && fc.au.HasPermission(rbac.PermissionRedemptionsApprove) {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("cannot request approval for your own reward"))
	}
	balance, err := s.store.GetMemberBalance(ctx, child.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if balance < reward.CostStars {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("insufficient balance"))
	}
	if !s.rewardAvailableNow(ctx, reward, child.ID, balance, time.Now()) {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("reward not available"))
	}
	if reward.ApprovalRequired {
		id, err := s.store.CreateRedemption(ctx, fc.family.ID, child.ID, reward.ID, reward.CostStars, store.RedemptionPending, nil)
		if err != nil {
			return nil, mapStoreError(err)
		}
		red, _ := s.store.GetRedemptionByID(ctx, id)
		s.webhooks.Dispatch(ctx, "redemption.requested", map[string]any{
			"family_id": fc.family.ID, "child_member_id": child.ID, "reward_id": reward.ID,
			"stars_spent": reward.CostStars, "redemption_id": id,
		})
		s.notifyRedemptionRequested(ctx, red)
		return connect.NewResponse(&apiv1.RequestRedemptionResponse{
			StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Redemption requested"},
			Redemption:       toProtoRedemption(red),
		}), nil
	}
	createdBy := fc.member.ID
	var ledgerCreator *int
	if fc.au.HasPermission(rbac.PermissionStarsViewFamily) {
		ledgerCreator = &createdBy
	}
	neg := -reward.CostStars
	rewardID := reward.ID
	entryID, err := s.store.InsertLedgerEntry(ctx, store.StarLedgerRow{
		FamilyID: fc.family.ID, ChildMemberID: child.ID, Amount: neg,
		EntryType: store.LedgerTypeRedeem, RelatedRewardID: &rewardID, CreatedByMemberID: ledgerCreator,
	})
	if err != nil {
		return nil, mapStoreError(err)
	}
	id, err := s.store.CreateRedemption(ctx, fc.family.ID, child.ID, reward.ID, reward.CostStars, store.RedemptionApproved, &entryID)
	if err != nil {
		return nil, mapStoreError(err)
	}
	red, _ := s.store.GetRedemptionByID(ctx, id)
	s.webhooks.Dispatch(ctx, "redemption.resolved", map[string]any{
		"redemption_id": id, "status": store.RedemptionApproved, "resolved_by_member_id": createdBy,
	})
	return connect.NewResponse(&apiv1.RequestRedemptionResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Redemption approved"},
		Redemption:       toProtoRedemption(red),
	}), nil
}

func (s *Server) ApproveRedemption(ctx context.Context, req *connect.Request[apiv1.ApproveRedemptionRequest]) (*connect.Response[apiv1.ApproveRedemptionResponse], error) {
	return s.resolveRedemption(ctx, int(req.Msg.RedemptionId), store.RedemptionApproved)
}

func (s *Server) RejectRedemption(ctx context.Context, req *connect.Request[apiv1.RejectRedemptionRequest]) (*connect.Response[apiv1.RejectRedemptionResponse], error) {
	res, err := s.resolveRedemption(ctx, int(req.Msg.RedemptionId), store.RedemptionRejected)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&apiv1.RejectRedemptionResponse{
		StandardResponse: res.Msg.StandardResponse,
		Redemption:       res.Msg.Redemption,
	}), nil
}

func (s *Server) resolveRedemption(ctx context.Context, redemptionID int, status string) (*connect.Response[apiv1.ApproveRedemptionResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionRedemptionsApprove); err != nil {
		return nil, err
	}
	red, err := s.store.GetRedemptionByID(ctx, redemptionID)
	if err != nil || red == nil || red.FamilyID != fc.family.ID || red.Status != store.RedemptionPending {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("redemption not found"))
	}
	if red.ChildMemberID == fc.member.ID {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("cannot approve or reject your own redemption"))
	}
	var entryID *int
	if status == store.RedemptionApproved {
		balance, err := s.store.GetMemberBalance(ctx, red.ChildMemberID)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		if balance < red.StarsSpent {
			return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("insufficient balance"))
		}
		createdBy := fc.member.ID
		rewardID := red.RewardID
		neg := -red.StarsSpent
		id, err := s.store.InsertLedgerEntry(ctx, store.StarLedgerRow{
			FamilyID: fc.family.ID, ChildMemberID: red.ChildMemberID, Amount: neg,
			EntryType: store.LedgerTypeRedeem, RelatedRewardID: &rewardID, CreatedByMemberID: &createdBy,
		})
		if err != nil {
			return nil, mapStoreError(err)
		}
		entryID = &id
	}
	if err := s.store.ResolveRedemption(ctx, red.ID, status, fc.member.ID, entryID); err != nil {
		return nil, mapStoreError(err)
	}
	red, _ = s.store.GetRedemptionByID(ctx, red.ID)
	s.webhooks.Dispatch(ctx, "redemption.resolved", map[string]any{
		"redemption_id": red.ID, "status": status, "resolved_by_member_id": fc.member.ID,
	})
	return connect.NewResponse(&apiv1.ApproveRedemptionResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Redemption updated"},
		Redemption:       toProtoRedemption(red),
	}), nil
}

func (s *Server) ListRedemptions(ctx context.Context, req *connect.Request[apiv1.ListRedemptionsRequest]) (*connect.Response[apiv1.ListRedemptionsResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if !fc.au.HasPermission(rbac.PermissionRedemptionsApprove) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("forbidden"))
	}
	rows, err := s.store.ListRedemptions(ctx, fc.family.ID, req.Msg.Status)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := &apiv1.ListRedemptionsResponse{}
	for i := range rows {
		out.Redemptions = append(out.Redemptions, toProtoRedemption(&rows[i]))
	}
	return connect.NewResponse(out), nil
}
