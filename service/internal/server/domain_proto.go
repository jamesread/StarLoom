package server

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/auth"
	"github.com/jamesread/starapp/service/internal/rbac"
	"github.com/jamesread/starapp/service/internal/store"
)

type familyContext struct {
	family *store.FamilyRow
	member *store.FamilyMemberRow
	au     *auth.AuthenticatedUser
}

func (s *Server) requirePermission(ctx context.Context, perm string) (*auth.AuthenticatedUser, error) {
	au, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if !au.HasPermission(perm) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("missing permission %s", perm))
	}
	return au, nil
}

func (s *Server) loadFamilyContext(ctx context.Context) (*familyContext, error) {
	au, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	member, err := s.store.GetMemberByAccountID(ctx, au.User.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if member == nil {
		return &familyContext{au: au}, nil
	}
	family, err := s.store.GetFamilyByID(ctx, member.FamilyID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return &familyContext{family: family, member: member, au: au}, nil
}

func (s *Server) requireFamilyContext(ctx context.Context) (*familyContext, error) {
	fc, err := s.loadFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if fc.family == nil || fc.member == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("family not set up"))
	}
	return fc, nil
}

func (s *Server) canViewMember(au *auth.AuthenticatedUser, caller, target *store.FamilyMemberRow) bool {
	if au.HasPermission(rbac.PermissionStarsViewFamily) {
		return caller != nil && target != nil && caller.FamilyID == target.FamilyID
	}
	if au.HasPermission(rbac.PermissionStarsViewOwn) {
		return caller != nil && target != nil && caller.ID == target.ID
	}
	return false
}

func toProtoFamily(f *store.FamilyRow) *apiv1.Family {
	if f == nil {
		return nil
	}
	return &apiv1.Family{Id: int32(f.ID), Name: f.Name, CreatedAt: f.CreatedAt}
}

func toProtoMember(m *store.FamilyMemberRow) *apiv1.FamilyMember {
	if m == nil {
		return nil
	}
	var accountID int32
	if m.UserAccountID != nil {
		accountID = int32(*m.UserAccountID)
	}
	starColor := m.StarColor
	if starColor == "" {
		starColor = store.DefaultMemberStarColor(m.ID)
	}
	return &apiv1.FamilyMember{
		Id:            int32(m.ID),
		FamilyId:      int32(m.FamilyID),
		UserAccountId: accountID,
		DisplayName:   m.DisplayName,
		Role:          m.Role,
		HasAvatar:     m.AvatarPath != "",
		CreatedAt:     m.CreatedAt,
		Username:      m.Username,
		StarColor:     starColor,
	}
}

func toProtoLedger(e *store.StarLedgerRow) *apiv1.StarLedgerEntry {
	if e == nil {
		return nil
	}
	var rewardID int32
	if e.RelatedRewardID != nil {
		rewardID = int32(*e.RelatedRewardID)
	}
	var createdBy int32
	if e.CreatedByMemberID != nil {
		createdBy = int32(*e.CreatedByMemberID)
	}
	return &apiv1.StarLedgerEntry{
		Id:                int32(e.ID),
		FamilyId:          int32(e.FamilyID),
		ChildMemberId:     int32(e.ChildMemberID),
		Amount:            int32(e.Amount),
		EntryType:         e.EntryType,
		Note:              e.Note,
		RelatedRewardId:   rewardID,
		CreatedByMemberId: createdBy,
		CreatedAt:         e.CreatedAt,
	}
}

func toProtoReward(r *store.RewardRow) *apiv1.Reward {
	if r == nil {
		return nil
	}
	return &apiv1.Reward{
		Id:               int32(r.ID),
		FamilyId:         int32(r.FamilyID),
		Title:            r.Title,
		Description:      r.Description,
		CostStars:        int32(r.CostStars),
		Active:           r.Active,
		ApprovalRequired: r.ApprovalRequired,
	}
}

func toProtoRedemption(r *store.RedemptionRow) *apiv1.Redemption {
	if r == nil {
		return nil
	}
	var resolvedBy int32
	if r.ResolvedByMemberID != nil {
		resolvedBy = int32(*r.ResolvedByMemberID)
	}
	return &apiv1.Redemption{
		Id:                 int32(r.ID),
		FamilyId:           int32(r.FamilyID),
		ChildMemberId:      int32(r.ChildMemberID),
		RewardId:           int32(r.RewardID),
		StarsSpent:         int32(r.StarsSpent),
		Status:             r.Status,
		CreatedAt:          r.CreatedAt,
		ResolvedAt:         r.ResolvedAt,
		ResolvedByMemberId: resolvedBy,
		FulfilledAt:        r.FulfilledAt,
		RewardTitle:        r.RewardTitle,
		ChildDisplayName:   r.ChildDisplayName,
	}
}
