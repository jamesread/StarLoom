package server

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/gen/starapp/api/v1/apiv1connect"
	"github.com/jamesread/starapp/service/internal/auth"
	"github.com/jamesread/starapp/service/internal/config"
	"github.com/jamesread/starapp/service/internal/store"
	"github.com/jamesread/starapp/service/internal/webhook"
)

var initTotal = promauto.NewCounter(prometheus.CounterOpts{
	Name: "starapp_init_total",
	Help: "Total Init RPC calls",
})

// Server implements StarAppServiceHandler.
type Server struct {
	cfg      *config.Config
	store    store.Store
	webhooks *webhook.Dispatcher
	log      *logrus.Logger
}

// New returns a Server wired to cfg and store.
func New(cfg *config.Config, st store.Store, hooks *webhook.Dispatcher, log *logrus.Logger) *Server {
	if log == nil {
		log = logrus.New()
	}
	if hooks == nil {
		hooks = &webhook.Dispatcher{Store: st}
	}
	return &Server{cfg: cfg, store: st, webhooks: hooks, log: log}
}

// Handler returns the Connect HTTP handler and mount path, wrapped with auth when layer is non-nil.
func (s *Server) Handler(layer *auth.Layer) (string, http.Handler) {
	path, h := apiv1connect.NewStarAppServiceHandler(s)
	if layer != nil {
		h = layer.WrapHandler(h)
	}
	return path, h
}

// Init returns SPA bootstrap metadata.
func (s *Server) Init(ctx context.Context, _ *connect.Request[apiv1.InitRequest]) (*connect.Response[apiv1.InitResponse], error) {
	initTotal.Inc()

	if err := s.store.Ping(ctx); err != nil {
		s.log.WithError(err).Error("store ping")
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(s.initShell(ctx)), nil
}
