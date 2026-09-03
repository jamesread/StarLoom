package server

import (
	"context"
	"net/http"

	"connectrpc.com/connect"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/apprise"
	"github.com/jamesread/starapp/service/internal/store"
)

func toProtoNotificationDelivery(row *store.NotificationDeliveryRow) *apiv1.NotificationDelivery {
	if row == nil {
		return nil
	}
	return &apiv1.NotificationDelivery{
		Id:                   int32(row.ID),
		RecipientMemberId:    int32(row.RecipientMemberID),
		RecipientDisplayName: row.RecipientDisplayName,
		NotificationType:     row.NotificationType,
		Title:                row.Title,
		Success:              row.Success,
		ErrorMessage:         row.ErrorMessage,
		SentAt:               row.SentAt,
	}
}

func (s *Server) ListNotificationDeliveries(ctx context.Context, req *connect.Request[apiv1.ListNotificationDeliveriesRequest]) (*connect.Response[apiv1.ListNotificationDeliveriesResponse], error) {
	limit := int(req.Msg.Limit)
	if limit <= 0 {
		limit = 30
	}
	if limit > 500 {
		limit = 500
	}
	rows, err := s.store.ListNotificationDeliveries(ctx, limit)
	if err != nil {
		s.log.WithError(err).Error("list notification deliveries")
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := &apiv1.ListNotificationDeliveriesResponse{}
	for i := range rows {
		out.Deliveries = append(out.Deliveries, toProtoNotificationDelivery(&rows[i]))
	}
	return connect.NewResponse(out), nil
}

func (s *Server) deliverApprise(ctx context.Context, client *http.Client, url string, payload apprise.Payload, familyID, recipientMemberID int, notificationType string) error {
	err := apprise.Notify(client, url, payload)
	row := store.NotificationDeliveryRow{
		FamilyID:          familyID,
		RecipientMemberID: recipientMemberID,
		NotificationType:  notificationType,
		Title:             payload.Title,
		Success:           err == nil,
	}
	if err != nil {
		row.ErrorMessage = err.Error()
	}
	if _, insertErr := s.store.InsertNotificationDelivery(ctx, row); insertErr != nil {
		s.log.WithError(insertErr).Warn("notification delivery history insert failed")
	}
	return err
}
