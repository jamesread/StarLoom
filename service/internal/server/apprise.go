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
	"github.com/jamesread/starapp/service/internal/store"
)

func (s *Server) SendMemberTestNotification(ctx context.Context, req *connect.Request[apiv1.SendMemberTestNotificationRequest]) (*connect.Response[apiv1.SendMemberTestNotificationResponse], error) {
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	member, err := s.testNotificationMember(ctx, fc, int(req.Msg.MemberId))
	if err != nil {
		return nil, err
	}
	return s.sendTestNotification(ctx, member)
}

func (s *Server) testNotificationMember(ctx context.Context, fc *familyContext, memberID int) (*store.FamilyMemberRow, error) {
	if memberID <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("member_id required"))
	}
	member, err := s.familyMemberOrNotFound(ctx, fc.family.ID, memberID)
	if err != nil {
		return nil, err
	}
	if !s.canViewMember(fc.au, fc.member, member) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("forbidden"))
	}
	return member, nil
}

func (s *Server) sendTestNotification(ctx context.Context, member *store.FamilyMemberRow) (*connect.Response[apiv1.SendMemberTestNotificationResponse], error) {
	url := strings.TrimSpace(s.stringCvar(ctx, cvar.KeyAppriseURL))
	if url == "" {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("apprise URL is not configured (Control Panel → Settings → Notifications)"))
	}
	tag := apprise.PersonTag(member.ID)
	client := &http.Client{Timeout: 15 * time.Second}
	payload := apprise.Payload{
		Title: "StarLoom test notification",
		Body:  fmt.Sprintf("Test notification for %s (tag %s).", member.DisplayName, tag),
		Type:  "info",
		Tag:   tag,
	}
	if err := s.deliverApprise(ctx, client, url, payload, member.FamilyID, member.ID, store.NotificationTypeTest); err != nil {
		s.log.WithError(err).WithField("member_id", member.ID).Warn("apprise: member test notification failed")
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("apprise notify failed: %w", err))
	}
	return connect.NewResponse(&apiv1.SendMemberTestNotificationResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Test notification sent"},
		Tag:              tag,
	}), nil
}

func (s *Server) notifyRedemptionRequested(ctx context.Context, red *store.RedemptionRow) {
	if red == nil {
		return
	}
	url := strings.TrimSpace(s.stringCvar(ctx, cvar.KeyAppriseURL))
	if url == "" {
		return
	}
	parentIDs, err := s.familyParentIDs(ctx, red.FamilyID)
	if err != nil || len(parentIDs) == 0 {
		return
	}
	body := apprise.RenderRedemptionMessage(
		s.stringCvar(ctx, cvar.KeyAppriseRedemptionMessage),
		apprise.RedemptionPlaceholders{
			ApprovalURL:   apprise.ApprovalURL(s.stringCvar(ctx, cvar.KeyExternalBaseURL), red.ID),
			RequestorName: red.ChildDisplayName,
			RewardName:    red.RewardTitle,
			Stars:         red.StarsSpent,
			RedemptionID:  red.ID,
			RequestorID:   red.ChildMemberID,
		},
	)
	s.dispatchParentApprise(url, red.FamilyID, parentIDs, body)
}

func (s *Server) familyParentIDs(ctx context.Context, familyID int) ([]int, error) {
	members, err := s.store.ListMembersByFamily(ctx, familyID)
	if err != nil {
		s.log.WithError(err).Warn("apprise: list members for redemption notify")
		return nil, err
	}
	parentIDs := make([]int, 0)
	for i := range members {
		if members[i].Role == store.MemberRoleParent {
			parentIDs = append(parentIDs, members[i].ID)
		}
	}
	return parentIDs, nil
}

func (s *Server) dispatchParentApprise(url string, familyID int, parentIDs []int, body string) {
	client := &http.Client{Timeout: 15 * time.Second}
	go func() {
		bg := context.Background()
		for _, personID := range parentIDs {
			if err := s.deliverApprise(bg, client, url, apprise.Payload{
				Title: "Reward approval requested", Body: body, Type: "info", Tag: apprise.PersonTag(personID),
			}, familyID, personID, store.NotificationTypeRedemptionRequested); err != nil {
				s.log.WithError(err).WithField("person_id", personID).Warn("apprise: redemption approval notify failed")
			}
		}
	}()
}
