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

func (s *Server) TestAppriseNotification(ctx context.Context, _ *connect.Request[apiv1.TestAppriseNotificationRequest]) (*connect.Response[apiv1.TestAppriseNotificationResponse], error) {
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	url := strings.TrimSpace(s.stringCvar(ctx, cvar.KeyAppriseURL))
	if url == "" {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("Apprise URL is not configured (Control Panel → Settings → Notifications)"))
	}
	tag := apprise.PersonTag(fc.member.ID)
	client := &http.Client{Timeout: 15 * time.Second}
	payload := apprise.Payload{
		Title: "StarLoom test notification",
		Body:  fmt.Sprintf("Test notification for %s (tag %s).", fc.member.DisplayName, tag),
		Type:  "info",
		Tag:   tag,
	}
	if err := s.deliverApprise(ctx, client, url, payload, fc.family.ID, fc.member.ID, store.NotificationTypeTest); err != nil {
		s.log.WithError(err).Warn("apprise: test notification failed")
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("apprise notify failed: %w", err))
	}
	return connect.NewResponse(&apiv1.TestAppriseNotificationResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Test notification sent"},
		Tag:              tag,
	}), nil
}

func (s *Server) SendUserTestNotification(ctx context.Context, req *connect.Request[apiv1.SendUserTestNotificationRequest]) (*connect.Response[apiv1.SendUserTestNotificationResponse], error) {
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	userID := int(req.Msg.UserId)
	if userID <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user_id required"))
	}
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return nil, mapStoreError(err)
	}
	if user == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found"))
	}
	member, err := s.store.GetMemberByAccountID(ctx, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if member == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("user is not linked to a family person"))
	}
	url := strings.TrimSpace(s.stringCvar(ctx, cvar.KeyAppriseURL))
	if url == "" {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("Apprise URL is not configured (Control Panel → Settings → Notifications)"))
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
		s.log.WithError(err).WithField("user_id", userID).Warn("apprise: admin user test notification failed")
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("apprise notify failed: %w", err))
	}
	return connect.NewResponse(&apiv1.SendUserTestNotificationResponse{
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

	members, err := s.store.ListMembersByFamily(ctx, red.FamilyID)
	if err != nil {
		s.log.WithError(err).Warn("apprise: list members for redemption notify")
		return
	}
	parentIDs := make([]int, 0)
	for i := range members {
		if members[i].Role == store.MemberRoleParent {
			parentIDs = append(parentIDs, members[i].ID)
		}
	}
	if len(parentIDs) == 0 {
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

	notifyURL := url
	familyID := red.FamilyID
	client := &http.Client{Timeout: 15 * time.Second}
	go func() {
		bg := context.Background()
		for _, personID := range parentIDs {
			if err := s.deliverApprise(bg, client, notifyURL, apprise.Payload{
				Title: "Reward approval requested",
				Body:  body,
				Type:  "info",
				Tag:   apprise.PersonTag(personID),
			}, familyID, personID, store.NotificationTypeRedemptionRequested); err != nil {
				s.log.WithError(err).WithField("person_id", personID).Warn("apprise: redemption approval notify failed")
			}
		}
	}()
}
