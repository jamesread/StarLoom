package server

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/store"
	"github.com/jamesread/starapp/service/internal/webhook"
)

func toProtoWebhook(w *store.WebhookTargetRow) *apiv1.Webhook {
	if w == nil {
		return nil
	}
	return &apiv1.Webhook{
		Id:      int32(w.ID),
		Url:     w.URL,
		Events:  append([]string(nil), w.Events...),
		Enabled: w.Enabled,
		Created: w.Created,
		Updated: w.Updated,
	}
}

func validateCreateWebhookInput(req *apiv1.CreateWebhookRequest) (url string, events []string, err error) {
	url, urlErr := webhook.NormalizeURL(req.Url)
	if urlErr != nil {
		return "", nil, urlErr
	}
	events, eventErr := webhook.NormalizeEvents(req.Events)
	if eventErr != nil {
		return "", nil, eventErr
	}
	if strings.TrimSpace(req.Secret) == "" {
		return "", nil, fmt.Errorf("secret required")
	}
	return url, events, nil
}

func validateUpdateWebhookInput(req *apiv1.UpdateWebhookRequest) (url string, events []string, err error) {
	url = req.Url
	if url != "" {
		url, err = webhook.NormalizeURL(url)
		if err != nil {
			return "", nil, err
		}
	}
	events, err = webhook.NormalizeEvents(req.Events)
	return url, events, err
}

func (s *Server) ListWebhooks(ctx context.Context, _ *connect.Request[apiv1.ListWebhooksRequest]) (*connect.Response[apiv1.ListWebhooksResponse], error) {
	rows, err := s.store.ListWebhookTargets(ctx)
	if err != nil {
		s.log.WithError(err).Error("list webhooks")
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := &apiv1.ListWebhooksResponse{Events: append([]string(nil), webhook.SupportedEvents...)}
	for _, w := range rows {
		out.Webhooks = append(out.Webhooks, toProtoWebhook(&w))
	}
	return connect.NewResponse(out), nil
}

func (s *Server) CreateWebhook(ctx context.Context, req *connect.Request[apiv1.CreateWebhookRequest]) (*connect.Response[apiv1.CreateWebhookResponse], error) {
	url, events, valErr := validateCreateWebhookInput(req.Msg)
	if valErr != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, valErr)
	}
	id, err := s.store.CreateWebhookTarget(ctx, url, req.Msg.Secret, events, req.Msg.Enabled)
	if err != nil {
		s.log.WithError(err).Error("create webhook")
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	w, _ := s.store.FindWebhookTarget(ctx, id)
	return connect.NewResponse(&apiv1.CreateWebhookResponse{Webhook: toProtoWebhook(w)}), nil
}

func (s *Server) UpdateWebhook(ctx context.Context, req *connect.Request[apiv1.UpdateWebhookRequest]) (*connect.Response[apiv1.Webhook], error) {
	url, events, valErr := validateUpdateWebhookInput(req.Msg)
	if valErr != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, valErr)
	}
	if updErr := s.store.UpdateWebhookTarget(ctx, int(req.Msg.Id), url, req.Msg.Secret, events, req.Msg.Enabled, false); updErr != nil {
		s.log.WithError(updErr).Error("update webhook")
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("webhook not found"))
	}
	w, _ := s.store.FindWebhookTarget(ctx, int(req.Msg.Id))
	return connect.NewResponse(toProtoWebhook(w)), nil
}

func (s *Server) DeleteWebhook(ctx context.Context, req *connect.Request[apiv1.DeleteWebhookRequest]) (*connect.Response[apiv1.DeleteWebhookResponse], error) {
	if err := s.store.DeleteWebhookTarget(ctx, int(req.Msg.Id)); err != nil {
		s.log.WithError(err).Error("delete webhook")
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("webhook not found"))
	}
	return connect.NewResponse(&apiv1.DeleteWebhookResponse{}), nil
}
