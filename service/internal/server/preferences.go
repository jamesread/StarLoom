package server

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/i18n"
)

func (s *Server) GetUserPreferences(ctx context.Context, _ *connect.Request[apiv1.GetUserPreferencesRequest]) (*connect.Response[apiv1.GetUserPreferencesResponse], error) {
	au, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	prefs, err := s.store.GetUserPreferences(ctx, au.User.ID)
	if err != nil {
		s.log.WithError(err).Error("get user preferences")
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&apiv1.GetUserPreferencesResponse{
		Language:           prefs.Language,
		AvailableLanguages: i18n.AvailableLanguageCodes(),
		SidebarEnabled:     proto.Bool(prefs.SidebarEnabled),
		ThemeToggleEnabled: proto.Bool(prefs.ThemeToggleEnabled),
	}), nil
}

func (s *Server) SaveUserPreferences(ctx context.Context, req *connect.Request[apiv1.SaveUserPreferencesRequest]) (*connect.Response[apiv1.SaveUserPreferencesResponse], error) {
	au, err := s.requireWrite(ctx)
	if err != nil {
		return nil, err
	}
	language := strings.TrimSpace(req.Msg.Language)
	if !i18n.IsSupportedLanguage(language) {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unsupported language: %s", language))
	}
	if err := s.store.SaveUserPreferences(ctx, au.User.ID, language, req.Msg.SidebarEnabled, req.Msg.ThemeToggleEnabled); err != nil {
		s.log.WithError(err).Error("save user preferences")
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&apiv1.SaveUserPreferencesResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Preferences saved"},
		Username:         au.User.Username,
	}), nil
}
