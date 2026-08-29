package server

import (
	"context"
	"strings"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/buildinfo"
	"github.com/jamesread/starapp/service/internal/cvar"
	"github.com/jamesread/starapp/service/internal/webhook"
)

func (s *Server) showVersionNumber(ctx context.Context) bool {
	return s.boolCvar(ctx, cvar.KeyShowVersionNumber, true)
}

func (s *Server) showNewVersions(ctx context.Context) bool {
	return s.boolCvar(ctx, cvar.KeyShowNewVersions, false)
}

func (s *Server) availableVersion(_ context.Context) string {
	// Update check not wired yet; return empty until a newer release is detected.
	return ""
}

func (s *Server) initShell(ctx context.Context) *apiv1.InitResponse {
	title := s.siteTitle(ctx)
	showVersion := s.showVersionNumber(ctx)

	currentVersion := ""
	availableVersion := ""
	showNewVersions := false
	if showVersion {
		currentVersion = buildinfo.Version
		if s.showNewVersions(ctx) {
			availableVersion = strings.TrimSpace(s.availableVersion(ctx))
			if availableVersion != "" && !isNonUpdateVersion(availableVersion) {
				showNewVersions = true
			}
		}
	}

	return &apiv1.InitResponse{
		ShowFooter:        s.showFooter(ctx),
		ShowNewVersions:   showNewVersions,
		AvailableVersion:  availableVersion,
		CurrentVersion:    currentVersion,
		PageTitle:         title,
		ShowVersionNumber: showVersion,
		SiteTitle:         title,
		Features: &apiv1.Features{
			RedemptionApprovalDefault: s.redemptionApprovalDefault(ctx),
		},
		WebhookEvents: append([]string(nil), webhook.SupportedEvents...),
	}
}

func isNonUpdateVersion(v string) bool {
	switch strings.ToLower(v) {
	case "", "none", "?":
		return true
	}
	return strings.HasPrefix(strings.ToLower(v), "you-are-using")
}
