package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/authn"
	"connectrpc.com/connect"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/gen/starapp/api/v1/apiv1connect"
	"github.com/jamesread/starapp/service/internal/auth"
	"github.com/jamesread/starapp/service/internal/config"
	"github.com/jamesread/starapp/service/internal/cvar"
	"github.com/jamesread/starapp/service/internal/rbac"
	"github.com/jamesread/starapp/service/internal/store"
)

func superuserCtx(ctx context.Context) context.Context {
	return authn.SetInfo(ctx, &auth.AuthenticatedUser{
		User: &store.UserAccountRow{ID: 1, Username: "admin"},
		RBAC: &rbac.EffectiveRBAC{IsSuperuser: true, Permissions: map[string]bool{}},
	})
}

func TestInit(t *testing.T) {
	t.Setenv("STARAPP_DEV_DISABLE_AUTH", "true")
	cfg := &config.Config{ShowFooter: true}
	st := store.OpenMemory()
	require.NoError(t, EnsureDefaultCvars(context.Background(), st, "StarApp", true))

	svc := New(cfg, st, nil, logrus.New())
	authLayer := auth.NewLayer(st, nil, logrus.New())

	mux := http.NewServeMux()
	path, handler := svc.Handler(authLayer)
	mux.Handle(path, handler)
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	client := apiv1connect.NewStarAppServiceClient(ts.Client(), ts.URL)

	res, err := client.Init(context.Background(), connect.NewRequest(&apiv1.InitRequest{}))
	require.NoError(t, err)
	require.Equal(t, "StarApp", res.Msg.SiteTitle)
	require.True(t, res.Msg.ShowFooter)
	require.NotEmpty(t, res.Msg.CurrentVersion)
	require.True(t, res.Msg.Features.RedemptionApprovalDefault)
	require.Contains(t, res.Msg.WebhookEvents, "stars.awarded")
}

func TestListAndUpdateCvar(t *testing.T) {
	st := store.OpenMemory()
	ctx := superuserCtx(context.Background())
	require.NoError(t, EnsureDefaultCvars(ctx, st, "StarApp", true))

	svc := New(&config.Config{}, st, nil, logrus.New())

	list, err := svc.ListCvars(ctx, connect.NewRequest(&apiv1.ListCvarsRequest{}))
	require.NoError(t, err)
	require.NotEmpty(t, list.Msg.Cvars)

	_, err = svc.UpdateCvar(ctx, connect.NewRequest(&apiv1.UpdateCvarRequest{
		Key:         cvar.KeySiteTitle,
		ValueString: "Our Stars",
	}))
	require.NoError(t, err)

	init, err := svc.Init(ctx, connect.NewRequest(&apiv1.InitRequest{}))
	require.NoError(t, err)
	require.Equal(t, "Our Stars", init.Msg.SiteTitle)
}

func TestWebhookCRUD(t *testing.T) {
	st := store.OpenMemory()
	ctx := superuserCtx(context.Background())
	svc := New(&config.Config{}, st, nil, logrus.New())

	created, err := svc.CreateWebhook(ctx, connect.NewRequest(&apiv1.CreateWebhookRequest{
		Url:     "https://example.com/hook",
		Secret:  "s3cret",
		Events:  []string{"stars.awarded"},
		Enabled: true,
	}))
	require.NoError(t, err)
	require.Equal(t, "https://example.com/hook", created.Msg.Webhook.Url)

	list, err := svc.ListWebhooks(ctx, connect.NewRequest(&apiv1.ListWebhooksRequest{}))
	require.NoError(t, err)
	require.Len(t, list.Msg.Webhooks, 1)
	require.Contains(t, list.Msg.Events, "redemption.requested")

	_, err = svc.DeleteWebhook(ctx, connect.NewRequest(&apiv1.DeleteWebhookRequest{Id: created.Msg.Webhook.Id}))
	require.NoError(t, err)
}
