package server

import (
	"context"
	"strings"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/buildinfo"
	"github.com/jamesread/starapp/service/internal/cvar"
	"github.com/jamesread/starapp/service/internal/themes"
	"github.com/jamesread/starapp/service/internal/webhook"
)

func (s *Server) showVersionNumber(ctx context.Context) bool {
	return s.boolCvar(ctx, cvar.KeyShowVersionNumber, true)
}

func (s *Server) showNewVersions(ctx context.Context) bool {
	return s.boolCvar(ctx, cvar.KeyShowNewVersions, true)
}

func (s *Server) availableVersion(_ context.Context) string {
	// Update check not wired yet; return empty until a newer release is detected.
	return ""
}

func (s *Server) availableThemes(_ context.Context) []string {
	webuiDir := ""
	if s.cfg != nil {
		webuiDir = s.cfg.WebUIDir
	}
	return themes.AvailableNames(webuiDir)
}

func (s *Server) themeColorSchemeSwitcherEnabled(ctx context.Context) bool {
	return s.boolCvar(ctx, cvar.KeyThemeColorSchemeSwitcherEnabled, false)
}

func (s *Server) themeName(ctx context.Context) string {
	row, err := s.store.FindCvar(ctx, cvar.KeyThemeName)
	if err != nil || row == nil {
		return ""
	}
	return strings.TrimSpace(row.ValueString)
}

func (s *Server) themeControl(ctx context.Context) string {
	row, err := s.store.FindCvar(ctx, cvar.KeyThemeControl)
	if err != nil || row == nil || strings.TrimSpace(row.ValueString) == "" {
		return cvar.ThemeControlUser
	}
	value := strings.TrimSpace(row.ValueString)
	if value == cvar.ThemeControlSystem {
		return cvar.ThemeControlSystem
	}
	return cvar.ThemeControlUser
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
			RedemptionApprovalDefault:       s.redemptionApprovalDefault(ctx),
			ThemeColorSchemeSwitcherEnabled: s.themeColorSchemeSwitcherEnabled(ctx),
			ThemeName:                       s.themeName(ctx),
			ThemeControl:                    s.themeControl(ctx),
			AvailableThemes:                 s.availableThemes(ctx),
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
