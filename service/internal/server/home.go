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
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionStarsViewFamily); err != nil {
		return nil, err
	}
	members, err := s.store.ListMembersByFamily(ctx, fc.family.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	pending, _ := s.store.CountPendingRedemptions(ctx, fc.family.ID)
	out := &apiv1.GetParentHomeSummaryResponse{
		Family:             toProtoFamily(fc.family),
		PendingRedemptions: int32(pending),
	}
	for i := range members {
		if !isFamilyStarMember(&members[i], fc.family.ID) {
			continue
		}
		balance, _ := s.store.GetMemberBalance(ctx, members[i].ID)
		last, _ := s.store.GetLastAward(ctx, members[i].ID)
		progress, err := s.memberTodayStarChartProgress(ctx, fc, members[i].ID)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		out.Children = append(out.Children, &apiv1.ChildHomeSummary{
			Member:                 toProtoMember(&members[i]),
			Balance:                int32(balance),
			LastAward:              toProtoLedger(last),
			TodayStarChartProgress: progress,
		})
	}
	if fc.member == nil {
		return connect.NewResponse(out), nil
	}
	chores, err := s.listTodaysChores(ctx, fc, fc.member.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out.TodaysChores = chores
	return connect.NewResponse(out), nil
}

func (s *Server) GetChildHomeSummary(ctx context.Context, _ *connect.Request[apiv1.GetChildHomeSummaryRequest]) (*connect.Response[apiv1.GetChildHomeSummaryResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionStarsViewOwn); err != nil {
		return nil, err
	}
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
	pendingRewardIDs := map[int]struct{}{}
	out := &apiv1.GetChildHomeSummaryResponse{
		Member:  toProtoMember(fc.member),
		Balance: int32(balance),
	}
	for i := range pendingRows {
		if pendingRows[i].ChildMemberID != fc.member.ID {
			continue
		}
		pendingRewardIDs[pendingRows[i].RewardID] = struct{}{}
		out.PendingRewardIds = append(out.PendingRewardIds, int32(pendingRows[i].RewardID))
	}
	for i := range ledger {
		if ledger[i].EntryType == store.LedgerTypeAward || ledger[i].Amount > 0 {
			out.RecentAwards = append(out.RecentAwards, toProtoLedger(&ledger[i]))
		}
	}
	rewardIncluded := map[int]bool{}
	now := time.Now()
	for i := range rewards {
		if !rewards[i].Active {
			continue
		}
		out.Rewards = append(out.Rewards, toProtoReward(&rewards[i]))
		rewardIncluded[rewards[i].ID] = true
		if s.rewardUnavailableDueToSchedule(ctx, &rewards[i], fc.member.ID, balance, now) {
			out.UnavailableRewardIds = append(out.UnavailableRewardIds, int32(rewards[i].ID))
		}
	}
	for rewardID := range pendingRewardIDs {
		if rewardIncluded[rewardID] {
			continue
		}
		reward, err := s.store.GetRewardByID(ctx, rewardID)
		if err != nil || reward == nil || !reward.Active {
			continue
		}
		out.Rewards = append(out.Rewards, toProtoReward(reward))
	}
	chores, err := s.listTodaysChores(ctx, fc, fc.member.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out.TodaysChores = chores
	return connect.NewResponse(out), nil
}
