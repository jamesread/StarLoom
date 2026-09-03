package server

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/cvar"
	"github.com/jamesread/starapp/service/internal/rbac"
	"github.com/jamesread/starapp/service/internal/store"
)

func (s *Server) applyMemberStarAdjustment(ctx context.Context, fc *familyContext, member *store.FamilyMemberRow, adjustment int, note string) (int, error) {
	if adjustment == 0 {
		balance, err := s.store.GetMemberBalance(ctx, member.ID)
		return balance, err
	}
	if !isFamilyStarMember(member, fc.family.ID) {
		return 0, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("member cannot hold stars"))
	}
	note = strings.TrimSpace(note)
	if note == "" {
		note = "Manual adjustment"
	}
	createdBy := fc.member.ID
	if adjustment > 0 {
		if _, err := s.requirePermission(ctx, rbac.PermissionStarsAward); err != nil {
			return 0, err
		}
		if adjustment > cvar.MaxAwardStars {
			return 0, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("adjustment must be between -%d and %d", cvar.MaxAwardStars, cvar.MaxAwardStars))
		}
		_, err := s.store.InsertLedgerEntry(ctx, store.StarLedgerRow{
			FamilyID: fc.family.ID, ChildMemberID: member.ID, Amount: adjustment,
			EntryType: store.LedgerTypeAward, Note: note, CreatedByMemberID: &createdBy,
		})
		if err != nil {
			return 0, mapStoreError(err)
		}
		s.webhooks.Dispatch(ctx, "stars.awarded", map[string]any{
			"family_id": fc.family.ID, "child_member_id": member.ID, "amount": adjustment,
			"note": note, "created_by_member_id": createdBy,
		})
	} else {
		if _, err := s.requirePermission(ctx, rbac.PermissionStarsRevoke); err != nil {
			return 0, err
		}
		amount := -adjustment
		if amount > cvar.MaxAwardStars {
			return 0, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("adjustment must be between -%d and %d", cvar.MaxAwardStars, cvar.MaxAwardStars))
		}
		balance, err := s.store.GetMemberBalance(ctx, member.ID)
		if err != nil {
			return 0, connect.NewError(connect.CodeInternal, err)
		}
		if amount > balance {
			amount = balance
		}
		if amount == 0 {
			return balance, nil
		}
		neg := -amount
		_, err = s.store.InsertLedgerEntry(ctx, store.StarLedgerRow{
			FamilyID: fc.family.ID, ChildMemberID: member.ID, Amount: neg,
			EntryType: store.LedgerTypeRevoke, Note: note, CreatedByMemberID: &createdBy,
		})
		if err != nil {
			return 0, mapStoreError(err)
		}
	}
	balance, err := s.store.GetMemberBalance(ctx, member.ID)
	return balance, err
}

func (s *Server) AwardStars(ctx context.Context, req *connect.Request[apiv1.AwardStarsRequest]) (*connect.Response[apiv1.AwardStarsResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionStarsAward); err != nil {
		return nil, err
	}
	member, err := s.store.GetMemberByID(ctx, int(req.Msg.ChildMemberId))
	if err != nil || !isFamilyStarMember(member, fc.family.ID) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("member not found"))
	}
	if member.ID == fc.member.ID {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("cannot award stars to yourself"))
	}
	amount := int(req.Msg.Amount)
	if amount <= 0 {
		amount = s.defaultAwardStars(ctx)
	}
	if amount <= 0 || amount > cvar.MaxAwardStars {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("amount must be 1-%d", cvar.MaxAwardStars))
	}
	createdBy := fc.member.ID
	entryID, err := s.store.InsertLedgerEntry(ctx, store.StarLedgerRow{
		FamilyID: fc.family.ID, ChildMemberID: member.ID, Amount: amount,
		EntryType: store.LedgerTypeAward, Note: req.Msg.Note, CreatedByMemberID: &createdBy,
	})
	if err != nil {
		return nil, mapStoreError(err)
	}
	balance, _ := s.store.GetMemberBalance(ctx, member.ID)
	ledger := &store.StarLedgerRow{
		ID: entryID, FamilyID: fc.family.ID, ChildMemberID: member.ID, Amount: amount,
		EntryType: store.LedgerTypeAward, Note: req.Msg.Note, CreatedByMemberID: &createdBy,
	}
	s.webhooks.Dispatch(ctx, "stars.awarded", map[string]any{
		"family_id": fc.family.ID, "child_member_id": member.ID, "amount": amount,
		"note": req.Msg.Note, "created_by_member_id": createdBy,
	})
	return connect.NewResponse(&apiv1.AwardStarsResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Stars awarded"},
		Entry:            toProtoLedger(ledger),
		NewBalance:       int32(balance),
	}), nil
}

func (s *Server) RevokeStars(ctx context.Context, req *connect.Request[apiv1.RevokeStarsRequest]) (*connect.Response[apiv1.RevokeStarsResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionStarsRevoke); err != nil {
		return nil, err
	}
	member, err := s.store.GetMemberByID(ctx, int(req.Msg.ChildMemberId))
	if err != nil || !isFamilyStarMember(member, fc.family.ID) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("member not found"))
	}
	amount := int(req.Msg.Amount)
	if amount <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("amount must be positive"))
	}
	balance, err := s.store.GetMemberBalance(ctx, member.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if amount > balance {
		amount = balance
	}
	if amount == 0 {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("nothing to revoke"))
	}
	createdBy := fc.member.ID
	neg := -amount
	entryID, err := s.store.InsertLedgerEntry(ctx, store.StarLedgerRow{
		FamilyID: fc.family.ID, ChildMemberID: member.ID, Amount: neg,
		EntryType: store.LedgerTypeRevoke, Note: req.Msg.Note, CreatedByMemberID: &createdBy,
	})
	if err != nil {
		return nil, mapStoreError(err)
	}
	balance, _ = s.store.GetMemberBalance(ctx, member.ID)
	return connect.NewResponse(&apiv1.RevokeStarsResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Stars revoked"},
		Entry:            toProtoLedger(&store.StarLedgerRow{ID: entryID, FamilyID: fc.family.ID, ChildMemberID: member.ID, Amount: neg, EntryType: store.LedgerTypeRevoke, Note: req.Msg.Note, CreatedByMemberID: &createdBy}),
		NewBalance:       int32(balance),
	}), nil
}

func (s *Server) GetMemberBalance(ctx context.Context, req *connect.Request[apiv1.GetMemberBalanceRequest]) (*connect.Response[apiv1.GetMemberBalanceResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	target, err := s.store.GetMemberByID(ctx, int(req.Msg.MemberId))
	if err != nil || target == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("member not found"))
	}
	if !s.canViewMember(fc.au, fc.member, target) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("forbidden"))
	}
	balance, err := s.store.GetMemberBalance(ctx, target.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&apiv1.GetMemberBalanceResponse{Balance: int32(balance)}), nil
}

func (s *Server) ListLedger(ctx context.Context, req *connect.Request[apiv1.ListLedgerRequest]) (*connect.Response[apiv1.ListLedgerResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	target, err := s.store.GetMemberByID(ctx, int(req.Msg.MemberId))
	if err != nil || target == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("member not found"))
	}
	if !s.canViewMember(fc.au, fc.member, target) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("forbidden"))
	}
	limit := int(req.Msg.Limit)
	rows, err := s.store.ListLedger(ctx, target.ID, limit)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := &apiv1.ListLedgerResponse{}
	for i := range rows {
		out.Entries = append(out.Entries, toProtoLedger(&rows[i]))
	}
	return connect.NewResponse(out), nil
}
