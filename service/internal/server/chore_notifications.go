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
	fc, memberID, err := s.authorizedNotificationMember(ctx, int(req.Msg.MemberId))
	if err != nil {
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
	fc, memberID, err := s.authorizedNotificationMember(ctx, int(req.Msg.MemberId))
	if err != nil {
		return nil, err
	}
	if err := s.saveChoreNotificationSubscriptions(ctx, fc, memberID, req.Msg.GetSubscriptions()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&apiv1.SaveMemberChoreNotificationSubscriptionsResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Notification subscriptions saved"},
	}), nil
}

func (s *Server) authorizedNotificationMember(ctx context.Context, memberID int) (*familyContext, int, error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, 0, err
	}
	if memberID <= 0 {
		return nil, 0, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("member_id required"))
	}
	target, err := s.familyMemberOrNotFound(ctx, fc.family.ID, memberID)
	if err != nil {
		return nil, 0, err
	}
	if err := s.authorizeMemberNotificationEdit(fc, target); err != nil {
		return nil, 0, err
	}
	return fc, memberID, nil
}

func (s *Server) authorizeMemberNotificationEdit(fc *familyContext, target *store.FamilyMemberRow) error {
	if fc.member != nil && target != nil && fc.member.ID == target.ID {
		return nil
	}
	return s.requireManageAndViewMember(fc, target)
}

func (s *Server) requireManageAndViewMember(fc *familyContext, target *store.FamilyMemberRow) error {
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
	memberIDs, choreByID, err := s.notificationLookupMaps(ctx, fc.family.ID)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	out := make([]store.ChoreNotificationSubscriptionRow, 0, len(subs))
	for _, sub := range subs {
		row, skip, err := normalizeSubscriptionRow(sub, fc.family.ID, memberIDs, choreByID, seen)
		if err != nil {
			return nil, err
		}
		if !skip {
			out = append(out, row)
		}
	}
	return out, nil
}

func (s *Server) notificationLookupMaps(ctx context.Context, familyID int) (map[int]bool, map[int]*store.ChoreWithAssignments, error) {
	memberIDs, err := s.familyMemberIDSet(ctx, familyID)
	if err != nil {
		return nil, nil, err
	}
	choreByID, err := s.choreByIDMap(ctx, familyID)
	if err != nil {
		return nil, nil, err
	}
	return memberIDs, choreByID, nil
}

func (s *Server) familyMemberIDSet(ctx context.Context, familyID int) (map[int]bool, error) {
	members, err := s.store.ListMembersByFamily(ctx, familyID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	memberIDs := map[int]bool{}
	for i := range members {
		if isFamilyStarMember(&members[i], familyID) {
			memberIDs[members[i].ID] = true
		}
	}
	return memberIDs, nil
}

func (s *Server) choreByIDMap(ctx context.Context, familyID int) (map[int]*store.ChoreWithAssignments, error) {
	chores, err := s.store.ListChores(ctx, familyID, 0, true)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	choreByID := map[int]*store.ChoreWithAssignments{}
	for i := range chores {
		choreByID[chores[i].Chore.ID] = &chores[i]
	}
	return choreByID, nil
}

func normalizeSubscriptionRow(
	sub *apiv1.ChoreNotificationSubscription,
	familyID int,
	memberIDs map[int]bool,
	choreByID map[int]*store.ChoreWithAssignments,
	seen map[string]bool,
) (store.ChoreNotificationSubscriptionRow, bool, error) {
	if sub == nil {
		return store.ChoreNotificationSubscriptionRow{}, true, nil
	}
	childID := int(sub.GetChildMemberId())
	choreID := int(sub.GetChoreId())
	if err := validateSubscriptionTargets(childID, choreID, familyID, memberIDs, choreByID); err != nil {
		return store.ChoreNotificationSubscriptionRow{}, false, err
	}
	key := subscriptionKey(childID, choreID)
	if seen[key] {
		return store.ChoreNotificationSubscriptionRow{}, true, nil
	}
	seen[key] = true
	return subscriptionRow(childID, choreID), false, nil
}

func validateSubscriptionTargets(
	childID, choreID, familyID int,
	memberIDs map[int]bool,
	choreByID map[int]*store.ChoreWithAssignments,
) error {
	if childID == 0 && choreID == 0 {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("subscription must specify a person, a chore, or both"))
	}
	if childID != 0 && !memberIDs[childID] {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown family member"))
	}
	return validateSubscriptionChore(childID, choreID, familyID, choreByID)
}

func validateSubscriptionChore(childID, choreID, familyID int, choreByID map[int]*store.ChoreWithAssignments) error {
	if choreID == 0 {
		return nil
	}
	cw, err := lookupSubscriptionChore(choreID, familyID, choreByID)
	if err != nil {
		return err
	}
	if childID != 0 && !assignmentExists(cw, childID) {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("chore is not assigned to that person"))
	}
	return nil
}

func lookupSubscriptionChore(choreID, familyID int, choreByID map[int]*store.ChoreWithAssignments) (*store.ChoreWithAssignments, error) {
	cw, ok := choreByID[choreID]
	if !ok || cw.Chore.FamilyID != familyID {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown chore"))
	}
	return cw, nil
}

func subscriptionRow(childID, choreID int) store.ChoreNotificationSubscriptionRow {
	row := store.ChoreNotificationSubscriptionRow{}
	if childID != 0 {
		row.ChildMemberID = &childID
	}
	if choreID != 0 {
		row.ChoreID = &choreID
	}
	return row
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
	payload, subscriberIDs, ok := s.choreCompletedNotifyPlan(ctx, familyID, childMemberID, choreID, choreTitle, starReward, date, completedByMemberID)
	if !ok {
		return
	}
	s.dispatchChoreCompleted(url, familyID, payload, subscriberIDs)
}

func (s *Server) choreCompletedNotifyPlan(ctx context.Context, familyID, childMemberID, choreID int, choreTitle string, starReward int, date string, completedByMemberID int) (apprise.Payload, []int, bool) {
	subscriberIDs, ok := s.choreCompletedSubscribers(ctx, familyID, childMemberID, choreID)
	if !ok {
		return apprise.Payload{}, nil, false
	}
	child, err := s.store.GetMemberByID(ctx, childMemberID)
	if err != nil || child == nil {
		s.log.WithError(err).Warn("apprise: chore completed child lookup")
		return apprise.Payload{}, nil, false
	}
	body := apprise.RenderChoreCompletedMessage(
		s.stringCvar(ctx, cvar.KeyAppriseChoreCompletedMessage),
		apprise.ChoreCompletedPlaceholders{
			ChildName: child.DisplayName, ChoreTitle: choreTitle, Stars: starReward, Date: date,
			CompletedByName: s.memberDisplayName(ctx, completedByMemberID),
		},
	)
	return apprise.Payload{Title: fmt.Sprintf("%s completed %s", child.DisplayName, choreTitle), Body: body, Type: "info"}, subscriberIDs, true
}

func (s *Server) choreCompletedSubscribers(ctx context.Context, familyID, childMemberID, choreID int) ([]int, bool) {
	subscriberIDs, err := s.store.MatchingChoreNotificationSubscribers(ctx, familyID, childMemberID, choreID)
	if err != nil {
		s.log.WithError(err).Warn("apprise: match chore notification subscribers")
		return nil, false
	}
	if len(subscriberIDs) == 0 {
		return nil, false
	}
	return subscriberIDs, true
}

func (s *Server) memberDisplayName(ctx context.Context, memberID int) string {
	if memberID == 0 {
		return ""
	}
	actor, err := s.store.GetMemberByID(ctx, memberID)
	if err != nil || actor == nil {
		return ""
	}
	return actor.DisplayName
}

func (s *Server) dispatchChoreCompleted(url string, familyID int, base apprise.Payload, subscriberIDs []int) {
	client := &http.Client{Timeout: 15 * time.Second}
	go func() {
		bg := context.Background()
		for _, personID := range subscriberIDs {
			payload := base
			payload.Tag = apprise.PersonTag(personID)
			if err := s.deliverApprise(bg, client, url, payload, familyID, personID, store.NotificationTypeChoreCompleted); err != nil {
				s.log.WithError(err).WithField("person_id", personID).Warn("apprise: chore completed notify failed")
			}
		}
	}()
}
