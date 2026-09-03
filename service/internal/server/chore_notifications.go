package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/apprise"
	"github.com/jamesread/starapp/service/internal/cvar"
	"github.com/jamesread/starapp/service/internal/rbac"
	"github.com/jamesread/starapp/service/internal/store"
)

func (s *Server) GetMyChoreNotificationSubscriptions(ctx context.Context, _ *connect.Request[apiv1.GetMyChoreNotificationSubscriptionsRequest]) (*connect.Response[apiv1.GetMyChoreNotificationSubscriptionsResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	subs, err := s.listChoreNotificationSubscriptions(ctx, fc, fc.member.ID)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&apiv1.GetMyChoreNotificationSubscriptionsResponse{Subscriptions: subs}), nil
}

func (s *Server) SaveMyChoreNotificationSubscriptions(ctx context.Context, req *connect.Request[apiv1.SaveMyChoreNotificationSubscriptionsRequest]) (*connect.Response[apiv1.SaveMyChoreNotificationSubscriptionsResponse], error) {
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.saveChoreNotificationSubscriptions(ctx, fc, fc.member.ID, req.Msg.GetSubscriptions()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&apiv1.SaveMyChoreNotificationSubscriptionsResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Notification subscriptions saved"},
	}), nil
}

func (s *Server) GetMemberChoreNotificationSubscriptions(ctx context.Context, req *connect.Request[apiv1.GetMemberChoreNotificationSubscriptionsRequest]) (*connect.Response[apiv1.GetMemberChoreNotificationSubscriptionsResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	memberID := int(req.Msg.MemberId)
	if memberID <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("member_id required"))
	}
	target, err := s.store.GetMemberByID(ctx, memberID)
	if err != nil || target == nil || !isFamilyStarMember(target, fc.family.ID) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("person not found"))
	}
	if err := s.authorizeMemberNotificationEdit(fc, target); err != nil {
		return nil, err
	}
	subs, err := s.listChoreNotificationSubscriptions(ctx, fc, memberID)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&apiv1.GetMemberChoreNotificationSubscriptionsResponse{
		Subscriptions:   subs,
		NotificationTag: apprise.PersonTag(memberID),
	}), nil
}

func (s *Server) SaveMemberChoreNotificationSubscriptions(ctx context.Context, req *connect.Request[apiv1.SaveMemberChoreNotificationSubscriptionsRequest]) (*connect.Response[apiv1.SaveMemberChoreNotificationSubscriptionsResponse], error) {
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	memberID := int(req.Msg.MemberId)
	if memberID <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("member_id required"))
	}
	target, err := s.store.GetMemberByID(ctx, memberID)
	if err != nil || target == nil || !isFamilyStarMember(target, fc.family.ID) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("person not found"))
	}
	if err := s.authorizeMemberNotificationEdit(fc, target); err != nil {
		return nil, err
	}
	if err := s.saveChoreNotificationSubscriptions(ctx, fc, memberID, req.Msg.GetSubscriptions()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&apiv1.SaveMemberChoreNotificationSubscriptionsResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Notification subscriptions saved"},
	}), nil
}

func (s *Server) authorizeMemberNotificationEdit(fc *familyContext, target *store.FamilyMemberRow) error {
	if fc.member != nil && target != nil && fc.member.ID == target.ID {
		return nil
	}
	if !fc.au.HasPermission(rbac.PermissionMembersManage) {
		return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("forbidden"))
	}
	if !s.canViewMember(fc.au, fc.member, target) {
		return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("forbidden"))
	}
	return nil
}

func (s *Server) listChoreNotificationSubscriptions(ctx context.Context, fc *familyContext, subscriberMemberID int) ([]*apiv1.ChoreNotificationSubscription, error) {
	rows, err := s.store.ListChoreNotificationSubscriptions(ctx, subscriberMemberID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	members, _ := s.store.ListMembersByFamily(ctx, fc.family.ID)
	chores, _ := s.store.ListChores(ctx, fc.family.ID, 0, true)
	memberName := memberNameByID(members)
	choreTitle := choreTitleByID(chores)
	out := make([]*apiv1.ChoreNotificationSubscription, 0, len(rows))
	for i := range rows {
		out = append(out, toProtoChoreNotificationSubscription(&rows[i], memberName, choreTitle))
	}
	return out, nil
}

func (s *Server) saveChoreNotificationSubscriptions(ctx context.Context, fc *familyContext, subscriberMemberID int, subs []*apiv1.ChoreNotificationSubscription) error {
	normalized, err := s.normalizeChoreNotificationSubscriptions(ctx, fc, subs)
	if err != nil {
		return err
	}
	if err := s.store.ReplaceChoreNotificationSubscriptions(ctx, fc.family.ID, subscriberMemberID, normalized); err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}

func (s *Server) normalizeChoreNotificationSubscriptions(ctx context.Context, fc *familyContext, subs []*apiv1.ChoreNotificationSubscription) ([]store.ChoreNotificationSubscriptionRow, error) {
	members, err := s.store.ListMembersByFamily(ctx, fc.family.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	memberIDs := map[int]bool{}
	for i := range members {
		if isFamilyStarMember(&members[i], fc.family.ID) {
			memberIDs[members[i].ID] = true
		}
	}
	chores, err := s.store.ListChores(ctx, fc.family.ID, 0, true)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	choreByID := map[int]*store.ChoreWithAssignments{}
	for i := range chores {
		choreByID[chores[i].Chore.ID] = &chores[i]
	}
	seen := map[string]bool{}
	out := make([]store.ChoreNotificationSubscriptionRow, 0, len(subs))
	for _, sub := range subs {
		if sub == nil {
			continue
		}
		childID := int(sub.GetChildMemberId())
		choreID := int(sub.GetChoreId())
		if childID == 0 && choreID == 0 {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("subscription must specify a person, a chore, or both"))
		}
		if childID != 0 && !memberIDs[childID] {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown family member"))
		}
		if choreID != 0 {
			cw, ok := choreByID[choreID]
			if !ok || cw.Chore.FamilyID != fc.family.ID {
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown chore"))
			}
			if childID != 0 && !assignmentExists(cw, childID) {
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("chore is not assigned to that person"))
			}
		}
		key := subscriptionKey(childID, choreID)
		if seen[key] {
			continue
		}
		seen[key] = true
		row := store.ChoreNotificationSubscriptionRow{}
		if childID != 0 {
			row.ChildMemberID = &childID
		}
		if choreID != 0 {
			row.ChoreID = &choreID
		}
		out = append(out, row)
	}
	return out, nil
}

func assignmentExists(cw *store.ChoreWithAssignments, childMemberID int) bool {
	if cw == nil {
		return false
	}
	for _, assign := range cw.Assignments {
		if assign.ChildMemberID == childMemberID {
			return true
		}
	}
	return false
}

func subscriptionKey(childID, choreID int) string {
	return fmt.Sprintf("%d:%d", childID, choreID)
}

func memberNameByID(members []store.FamilyMemberRow) map[int]string {
	out := map[int]string{}
	for i := range members {
		out[members[i].ID] = members[i].DisplayName
	}
	return out
}

func choreTitleByID(chores []store.ChoreWithAssignments) map[int]string {
	out := map[int]string{}
	for i := range chores {
		out[chores[i].Chore.ID] = chores[i].Chore.Title
	}
	return out
}

func toProtoChoreNotificationSubscription(row *store.ChoreNotificationSubscriptionRow, memberName, choreTitle map[int]string) *apiv1.ChoreNotificationSubscription {
	if row == nil {
		return nil
	}
	out := &apiv1.ChoreNotificationSubscription{}
	if row.ChildMemberID != nil {
		out.ChildMemberId = int32(*row.ChildMemberID)
		out.ChildDisplayName = memberName[*row.ChildMemberID]
	}
	if row.ChoreID != nil {
		out.ChoreId = int32(*row.ChoreID)
		out.ChoreTitle = choreTitle[*row.ChoreID]
	}
	return out
}

func (s *Server) notifyChoreCompleted(ctx context.Context, familyID, childMemberID, choreID int, choreTitle string, starReward int, date string, completedByMemberID int) {
	url := strings.TrimSpace(s.stringCvar(ctx, cvar.KeyAppriseURL))
	if url == "" {
		return
	}
	subscriberIDs, err := s.store.MatchingChoreNotificationSubscribers(ctx, familyID, childMemberID, choreID)
	if err != nil {
		s.log.WithError(err).Warn("apprise: match chore notification subscribers")
		return
	}
	if len(subscriberIDs) == 0 {
		return
	}
	child, err := s.store.GetMemberByID(ctx, childMemberID)
	if err != nil || child == nil {
		s.log.WithError(err).Warn("apprise: chore completed child lookup")
		return
	}
	completedByName := ""
	if completedByMemberID != 0 {
		if actor, err := s.store.GetMemberByID(ctx, completedByMemberID); err == nil && actor != nil {
			completedByName = actor.DisplayName
		}
	}
	body := apprise.RenderChoreCompletedMessage(
		s.stringCvar(ctx, cvar.KeyAppriseChoreCompletedMessage),
		apprise.ChoreCompletedPlaceholders{
			ChildName:       child.DisplayName,
			ChoreTitle:      choreTitle,
			Stars:           starReward,
			Date:            date,
			CompletedByName: completedByName,
		},
	)
	notifyURL := url
	client := &http.Client{Timeout: 15 * time.Second}
	childName := child.DisplayName
	go func() {
		bg := context.Background()
		for _, personID := range subscriberIDs {
			if err := s.deliverApprise(bg, client, notifyURL, apprise.Payload{
				Title: fmt.Sprintf("%s completed %s", childName, choreTitle),
				Body:  body,
				Type:  "info",
				Tag:   apprise.PersonTag(personID),
			}, familyID, personID, store.NotificationTypeChoreCompleted); err != nil {
				s.log.WithError(err).WithField("person_id", personID).Warn("apprise: chore completed notify failed")
			}
		}
	}()
}
