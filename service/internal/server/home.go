package server

import (
	"context"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/rbac"
	"github.com/jamesread/starapp/service/internal/store"
)

func (s *Server) GetParentHomeSummary(ctx context.Context, _ *connect.Request[apiv1.GetParentHomeSummaryRequest]) (*connect.Response[apiv1.GetParentHomeSummaryResponse], error) {
	fc, err := s.requireFamilyPermission(ctx, rbac.PermissionStarsViewFamily)
	if err != nil {
		return nil, err
	}
	out, err := s.buildParentHomeSummary(ctx, fc)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(out), nil
}

func (s *Server) buildParentHomeSummary(ctx context.Context, fc *familyContext) (*apiv1.GetParentHomeSummaryResponse, error) {
	members, err := s.store.ListMembersByFamily(ctx, fc.family.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	pending, _ := s.store.CountPendingRedemptions(ctx, fc.family.ID)
	out := &apiv1.GetParentHomeSummaryResponse{
		Family:             toProtoFamily(fc.family),
		PendingRedemptions: int32(pending),
	}
	children, err := s.parentHomeChildren(ctx, fc, members)
	if err != nil {
		return nil, err
	}
	out.Children = children
	if fc.member == nil {
		return out, nil
	}
	chores, err := s.listTodaysChores(ctx, fc, fc.member.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out.TodaysChores = chores
	return out, nil
}

func (s *Server) parentHomeChildren(ctx context.Context, fc *familyContext, members []store.FamilyMemberRow) ([]*apiv1.ChildHomeSummary, error) {
	out := make([]*apiv1.ChildHomeSummary, 0)
	for i := range members {
		if !isFamilyStarMember(&members[i], fc.family.ID) {
			continue
		}
		child, err := s.parentHomeChild(ctx, fc, &members[i])
		if err != nil {
			return nil, err
		}
		out = append(out, child)
	}
	return out, nil
}

func (s *Server) parentHomeChild(ctx context.Context, fc *familyContext, member *store.FamilyMemberRow) (*apiv1.ChildHomeSummary, error) {
	balance, _ := s.store.GetMemberBalance(ctx, member.ID)
	last, _ := s.store.GetLastAward(ctx, member.ID)
	progress, err := s.memberTodayStarChartProgress(ctx, fc, member.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return &apiv1.ChildHomeSummary{
		Member:                 toProtoMember(member),
		Balance:                int32(balance),
		LastAward:              toProtoLedger(last),
		TodayStarChartProgress: progress,
	}, nil
}

func (s *Server) GetChildHomeSummary(ctx context.Context, _ *connect.Request[apiv1.GetChildHomeSummaryRequest]) (*connect.Response[apiv1.GetChildHomeSummaryResponse], error) {
	fc, err := s.requireFamilyPermission(ctx, rbac.PermissionStarsViewOwn)
	if err != nil {
		return nil, err
	}
	out, err := s.buildChildHomeSummary(ctx, fc)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(out), nil
}

func (s *Server) buildChildHomeSummary(ctx context.Context, fc *familyContext) (*apiv1.GetChildHomeSummaryResponse, error) {
	data, err := s.loadChildHomeData(ctx, fc)
	if err != nil {
		return nil, err
	}
	out := &apiv1.GetChildHomeSummaryResponse{
		Member:  toProtoMember(fc.member),
		Balance: int32(data.balance),
	}
	fillChildPendingRewards(out, fc.member.ID, data.pendingRows)
	fillChildRecentAwards(out, data.ledger)
	fillChildRewards(ctx, s, out, fc, data.rewards, data.balance, data.pendingRewardIDs)
	chores, err := s.listTodaysChores(ctx, fc, fc.member.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out.TodaysChores = chores
	return out, nil
}

type childHomeData struct {
	pendingRewardIDs map[int]struct{}
	ledger           []store.StarLedgerRow
	rewards          []store.RewardRow
	pendingRows      []store.RedemptionRow
	balance          int
}

func (s *Server) loadChildHomeData(ctx context.Context, fc *familyContext) (*childHomeData, error) {
	balance, err := s.store.GetMemberBalance(ctx, fc.member.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	ledger, err := s.store.ListLedger(ctx, fc.member.ID, 20)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	rewards, err := s.store.ListRewards(ctx, fc.family.ID, false)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	pendingRows, err := s.store.ListRedemptions(ctx, fc.family.ID, store.RedemptionPending)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return &childHomeData{
		balance: balance, ledger: ledger, rewards: rewards, pendingRows: pendingRows,
		pendingRewardIDs: map[int]struct{}{},
	}, nil
}

func fillChildPendingRewards(out *apiv1.GetChildHomeSummaryResponse, memberID int, pendingRows []store.RedemptionRow) {
	for i := range pendingRows {
		if pendingRows[i].ChildMemberID != memberID {
			continue
		}
		out.PendingRewardIds = append(out.PendingRewardIds, int32(pendingRows[i].RewardID))
	}
}

func fillChildRecentAwards(out *apiv1.GetChildHomeSummaryResponse, ledger []store.StarLedgerRow) {
	for i := range ledger {
		if ledger[i].EntryType == store.LedgerTypeAward || ledger[i].Amount > 0 {
			out.RecentAwards = append(out.RecentAwards, toProtoLedger(&ledger[i]))
		}
	}
}

func fillChildRewards(
	ctx context.Context,
	s *Server,
	out *apiv1.GetChildHomeSummaryResponse,
	fc *familyContext,
	rewards []store.RewardRow,
	balance int,
	pendingRewardIDs map[int]struct{},
) {
	for i := range out.PendingRewardIds {
		pendingRewardIDs[int(out.PendingRewardIds[i])] = struct{}{}
	}
	included := map[int]bool{}
	now := time.Now()
	for i := range rewards {
		if !rewards[i].Active {
			continue
		}
		out.Rewards = append(out.Rewards, toProtoReward(&rewards[i]))
		included[rewards[i].ID] = true
		if s.rewardUnavailableDueToSchedule(ctx, &rewards[i], fc.member.ID, balance, now) {
			out.UnavailableRewardIds = append(out.UnavailableRewardIds, int32(rewards[i].ID))
		}
	}
	appendPendingRewards(ctx, s, out, pendingRewardIDs, included)
}

func appendPendingRewards(
	ctx context.Context,
	s *Server,
	out *apiv1.GetChildHomeSummaryResponse,
	pendingRewardIDs map[int]struct{},
	included map[int]bool,
) {
	for rewardID := range pendingRewardIDs {
		if included[rewardID] {
			continue
		}
		if reward := activeRewardOrNil(ctx, s, rewardID); reward != nil {
			out.Rewards = append(out.Rewards, toProtoReward(reward))
		}
	}
}

func activeRewardOrNil(ctx context.Context, s *Server, rewardID int) *store.RewardRow {
	reward, err := s.store.GetRewardByID(ctx, rewardID)
	if err != nil || reward == nil || !reward.Active {
		return nil
	}
	return reward
}
