package server

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/store"
)

func (s *Server) loadToggleChoreContext(
	ctx context.Context,
	fc *familyContext,
	choreID, childMemberID int,
	date string,
) (*store.ChoreWithAssignments, *store.ChoreAssignmentRow, error) {
	cw, err := s.activeFamilyChore(ctx, fc.family.ID, choreID)
	if err != nil {
		return nil, nil, err
	}
	assign, err := s.assignedChoreMember(ctx, fc.family.ID, choreID, childMemberID)
	if err != nil {
		return nil, nil, err
	}
	if err := s.ensureChoreSchedulable(ctx, fc.family.ID, cw.Chore.WeekdayMask, date); err != nil {
		return nil, nil, err
	}
	return cw, assign, nil
}

func (s *Server) activeFamilyChore(ctx context.Context, familyID, choreID int) (*store.ChoreWithAssignments, error) {
	cw, err := s.store.GetChoreByID(ctx, choreID)
	if err != nil || cw == nil || cw.Chore.FamilyID != familyID || !cw.Chore.Active {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("chore not found"))
	}
	return cw, nil
}

func (s *Server) assignedChoreMember(ctx context.Context, familyID, choreID, childMemberID int) (*store.ChoreAssignmentRow, error) {
	member, err := s.store.GetMemberByID(ctx, childMemberID)
	if err != nil || !isFamilyStarMember(member, familyID) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("member not found"))
	}
	assign, err := s.store.GetAssignment(ctx, choreID, childMemberID)
	if err != nil || assign == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("member not assigned to chore"))
	}
	return assign, nil
}

func (s *Server) ensureChoreSchedulable(ctx context.Context, familyID, weekdayMask int, date string) error {
	paused, _ := s.store.IsDatePaused(ctx, familyID, date)
	if paused {
		return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("chores are paused on this date"))
	}
	if !store.IsScheduledOnDate(weekdayMask, date) {
		return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("chore not scheduled on this date"))
	}
	return nil
}

func (s *Server) completeChoreAssignment(
	ctx context.Context,
	fc *familyContext,
	cw *store.ChoreWithAssignments,
	assign *store.ChoreAssignmentRow,
	childMemberID int,
	date string,
) (*connect.Response[apiv1.CompleteChoreResponse], error) {
	existing, err := s.store.GetCompletion(ctx, assign.ID, date)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if existing != nil {
		return choreToggleBalanceResponse(ctx, s, childMemberID, "Already completed")
	}
	createdBy := fc.member.ID
	entryID, err := s.store.InsertLedgerEntry(ctx, store.StarLedgerRow{
		FamilyID: fc.family.ID, ChildMemberID: childMemberID, Amount: cw.Chore.StarReward,
		EntryType: store.LedgerTypeAward, Note: fmt.Sprintf("Chore: %s", cw.Chore.Title),
		CreatedByMemberID: &createdBy,
	})
	if err != nil {
		return nil, mapStoreError(err)
	}
	if _, err := s.store.InsertChoreCompletion(ctx, assign.ID, date, entryID); err != nil {
		return nil, mapStoreError(err)
	}
	s.webhooks.Dispatch(ctx, "stars.awarded", map[string]any{
		"family_id": fc.family.ID, "child_member_id": childMemberID,
		"amount": cw.Chore.StarReward, "note": fmt.Sprintf("Chore: %s", cw.Chore.Title),
		"created_by_member_id": createdBy, "chore_id": cw.Chore.ID,
	})
	s.notifyChoreCompleted(ctx, fc.family.ID, childMemberID, cw.Chore.ID, cw.Chore.Title, cw.Chore.StarReward, date, createdBy)
	return choreToggleBalanceResponse(ctx, s, childMemberID, "Chore completed")
}

func (s *Server) uncompleteChoreAssignment(
	ctx context.Context,
	fc *familyContext,
	cw *store.ChoreWithAssignments,
	assign *store.ChoreAssignmentRow,
	childMemberID int,
	date string,
) (*connect.Response[apiv1.CompleteChoreResponse], error) {
	existing, err := s.store.GetCompletion(ctx, assign.ID, date)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if existing == nil {
		return choreToggleBalanceResponse(ctx, s, childMemberID, "Not completed")
	}
	createdBy := fc.member.ID
	neg := -cw.Chore.StarReward
	if _, err := s.store.InsertLedgerEntry(ctx, store.StarLedgerRow{
		FamilyID: fc.family.ID, ChildMemberID: childMemberID, Amount: neg,
		EntryType: store.LedgerTypeRevoke, Note: fmt.Sprintf("Undo chore: %s", cw.Chore.Title),
		CreatedByMemberID: &createdBy,
	}); err != nil {
		return nil, mapStoreError(err)
	}
	if err := s.store.DeleteChoreCompletion(ctx, existing.ID); err != nil {
		return nil, mapStoreError(err)
	}
	return choreToggleBalanceResponse(ctx, s, childMemberID, "Chore uncompleted")
}

func choreToggleBalanceResponse(ctx context.Context, s *Server, childMemberID int, message string) (*connect.Response[apiv1.CompleteChoreResponse], error) {
	balance, _ := s.store.GetMemberBalance(ctx, childMemberID)
	return connect.NewResponse(&apiv1.CompleteChoreResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: message},
		NewBalance:       int32(balance),
	}), nil
}
