package server

import (
	"context"

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
		out.Children = append(out.Children, &apiv1.ChildHomeSummary{
			Member:    toProtoMember(&members[i]),
			Balance:   int32(balance),
			LastAward: toProtoLedger(last),
		})
	}
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
	out := &apiv1.GetChildHomeSummaryResponse{
		Member:  toProtoMember(fc.member),
		Balance: int32(balance),
	}
	for i := range ledger {
		if ledger[i].EntryType == store.LedgerTypeAward || ledger[i].Amount > 0 {
			out.RecentAwards = append(out.RecentAwards, toProtoLedger(&ledger[i]))
		}
	}
	for i := range rewards {
		out.Rewards = append(out.Rewards, toProtoReward(&rewards[i]))
	}
	return connect.NewResponse(out), nil
}
