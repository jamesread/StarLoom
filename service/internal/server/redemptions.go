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
	fc, err := s.requireFamilyWritePermission(ctx, rbac.PermissionRedemptionsRequest)
	if err != nil {
		return nil, err
	}
	child, reward, cost, err := s.loadRedemptionRequest(ctx, fc, req.Msg)
	if err != nil {
		return nil, err
	}
	if reward.ApprovalRequired {
		return s.createPendingRedemption(ctx, fc, child, reward, cost)
	}
	return s.createImmediateRedemption(ctx, fc, child, reward, cost)
}

func (s *Server) loadRedemptionRequest(
	ctx context.Context,
	fc *familyContext,
	msg *apiv1.RequestRedemptionRequest,
) (*store.FamilyMemberRow, *store.RewardRow, int, error) {
	child, err := s.redemptionChild(ctx, fc, int(msg.ChildMemberId))
	if err != nil {
		return nil, nil, 0, err
	}
	reward, err := s.redemptionReward(ctx, fc, int(msg.RewardId), child)
	if err != nil {
		return nil, nil, 0, err
	}
	cost, err := s.redemptionCost(ctx, child, reward, int(msg.Quantity))
	if err != nil {
		return nil, nil, 0, err
	}
	return child, reward, cost, nil
}

func (s *Server) redemptionChild(ctx context.Context, fc *familyContext, childID int) (*store.FamilyMemberRow, error) {
	if childID == 0 {
		childID = fc.member.ID
	}
	child, err := s.store.GetMemberByID(ctx, childID)
	if err != nil || !isFamilyStarMember(child, fc.family.ID) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("member not found"))
	}
	if err := ensureCanRequestForChild(fc, child); err != nil {
		return nil, err
	}
	return child, nil
}

func ensureCanRequestForChild(fc *familyContext, child *store.FamilyMemberRow) error {
	if fc.au.HasPermission(rbac.PermissionStarsViewFamily) || child.ID == fc.member.ID {
		return nil
	}
	return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("forbidden"))
}

func (s *Server) redemptionReward(
	ctx context.Context,
	fc *familyContext,
	rewardID int,
	child *store.FamilyMemberRow,
) (*store.RewardRow, error) {
	reward, err := s.store.GetRewardByID(ctx, rewardID)
	if !isActiveFamilyReward(reward, err, fc.family.ID) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("reward not found"))
	}
	if err := ensureNotSelfApprovalRequest(fc, child, reward); err != nil {
		return nil, err
	}
	return reward, nil
}

func isActiveFamilyReward(reward *store.RewardRow, err error, familyID int) bool {
	return err == nil && reward != nil && reward.FamilyID == familyID && reward.Active
}

func ensureNotSelfApprovalRequest(fc *familyContext, child *store.FamilyMemberRow, reward *store.RewardRow) error {
	if reward.ApprovalRequired && child.ID == fc.member.ID && fc.au.HasPermission(rbac.PermissionRedemptionsApprove) {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("cannot request approval for your own reward"))
	}
	return nil
}

func (s *Server) redemptionCost(ctx context.Context, child *store.FamilyMemberRow, reward *store.RewardRow, qty int) (int, error) {
	cost, err := quantityCost(reward, qty)
	if err != nil {
		return 0, err
	}
	return s.ensureAffordableAvailable(ctx, child, reward, cost)
}

func quantityCost(reward *store.RewardRow, qty int) (int, error) {
	qty = max(qty, 1)
	if qty > 1 && !reward.MultiSpend {
		return 0, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("quantity must be 1 for this reward"))
	}
	return reward.CostStars * qty, nil
}

func (s *Server) ensureAffordableAvailable(ctx context.Context, child *store.FamilyMemberRow, reward *store.RewardRow, cost int) (int, error) {
	balance, err := s.store.GetMemberBalance(ctx, child.ID)
	if err != nil {
		return 0, connect.NewError(connect.CodeInternal, err)
	}
	if balance < cost {
		return 0, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("insufficient star balance (has %d, needs %d)", balance, cost))
	}
	if !s.rewardAvailableNow(ctx, reward, child.ID, balance, time.Now()) {
		return 0, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("reward not available"))
	}
	return cost, nil
}

func (s *Server) createPendingRedemption(
	ctx context.Context,
	fc *familyContext,
	child *store.FamilyMemberRow,
	reward *store.RewardRow,
	cost int,
) (*connect.Response[apiv1.RequestRedemptionResponse], error) {
	id, err := s.store.CreateRedemption(ctx, fc.family.ID, child.ID, reward.ID, cost, store.RedemptionPending, nil)
	if err != nil {
		return nil, mapStoreError(err)
	}
	red, _ := s.store.GetRedemptionByID(ctx, id)
	s.webhooks.Dispatch(ctx, "redemption.requested", map[string]any{
		"family_id": fc.family.ID, "child_member_id": child.ID, "reward_id": reward.ID,
		"stars_spent": cost, "redemption_id": id,
	})
	s.notifyRedemptionRequested(ctx, red)
	return connect.NewResponse(&apiv1.RequestRedemptionResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Redemption requested"},
		Redemption:       toProtoRedemption(red),
	}), nil
}

func (s *Server) createImmediateRedemption(
	ctx context.Context,
	fc *familyContext,
	child *store.FamilyMemberRow,
	reward *store.RewardRow,
	cost int,
) (*connect.Response[apiv1.RequestRedemptionResponse], error) {
	createdBy := fc.member.ID
	var ledgerCreator *int
	if fc.au.HasPermission(rbac.PermissionStarsViewFamily) {
		ledgerCreator = &createdBy
	}
	neg := -cost
	rewardID := reward.ID
	entryID, err := s.store.InsertLedgerEntry(ctx, store.StarLedgerRow{
		FamilyID: fc.family.ID, ChildMemberID: child.ID, Amount: neg,
		EntryType: store.LedgerTypeRedeem, RelatedRewardID: &rewardID, CreatedByMemberID: ledgerCreator,
	})
	if err != nil {
		return nil, mapStoreError(err)
	}
	id, err := s.store.CreateRedemption(ctx, fc.family.ID, child.ID, reward.ID, cost, store.RedemptionApproved, &entryID)
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
	fc, err := s.requireFamilyWritePermission(ctx, rbac.PermissionRedemptionsApprove)
	if err != nil {
		return nil, err
	}
	red, err := s.pendingFamilyRedemption(ctx, fc, redemptionID)
	if err != nil {
		return nil, err
	}
	entryID, err := s.ledgerEntryForApproval(ctx, fc, red, status)
	if err != nil {
		return nil, err
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

func (s *Server) pendingFamilyRedemption(ctx context.Context, fc *familyContext, redemptionID int) (*store.RedemptionRow, error) {
	red, err := s.store.GetRedemptionByID(ctx, redemptionID)
	if !isPendingFamilyRedemption(red, err, fc.family.ID) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("redemption not found"))
	}
	if red.ChildMemberID == fc.member.ID {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("cannot approve or reject your own redemption"))
	}
	return red, nil
}

func isPendingFamilyRedemption(red *store.RedemptionRow, err error, familyID int) bool {
	return err == nil && red != nil && red.FamilyID == familyID && red.Status == store.RedemptionPending
}

func (s *Server) ledgerEntryForApproval(ctx context.Context, fc *familyContext, red *store.RedemptionRow, status string) (*int, error) {
	if status != store.RedemptionApproved {
		return nil, nil
	}
	balance, err := s.store.GetMemberBalance(ctx, red.ChildMemberID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if balance < red.StarsSpent {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("insufficient star balance (has %d, needs %d)", balance, red.StarsSpent))
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
	return &id, nil
}

func (s *Server) ListRedemptions(ctx context.Context, req *connect.Request[apiv1.ListRedemptionsRequest]) (*connect.Response[apiv1.ListRedemptionsResponse], error) {
	fc, err := s.requireFamilyPermission(ctx, rbac.PermissionRedemptionsApprove)
	if err != nil {
		return nil, err
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
