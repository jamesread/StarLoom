package server

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/cvar"
	"github.com/jamesread/starapp/service/internal/store"
	"github.com/jamesread/starapp/service/internal/themes"
)

// EnsureDefaultCvars upserts catalog rows; metadata refresh only on conflict.
func EnsureDefaultCvars(ctx context.Context, st store.Store, siteTitle string, showFooter bool) error {
	for _, def := range cvar.Defaults(siteTitle, showFooter) {
		if err := st.InsertCvarIfMissing(ctx, store.CvarRow{
			Key: def.Key, MainType: def.MainType,
			Title: def.Title, Description: def.Description,
			Category: def.Category, Ordinal: def.Ordinal,
			ValueInt: def.ValueInt, ValueString: def.ValueString,
		}); err != nil {
			return err
		}
	}
	return nil
}

func toProtoCvar(row *store.CvarRow) *apiv1.Cvar {
	if row == nil {
		return nil
	}
	return &apiv1.Cvar{
		Key: row.Key, MainType: row.MainType,
		ValueInt: int32(row.ValueInt), ValueString: row.ValueString,
		Title: row.Title, Description: row.Description,
		Category: row.Category, Ordinal: int32(row.Ordinal),
	}
}

func appendProtoCvars(dst []*apiv1.Cvar, rows []store.CvarRow) []*apiv1.Cvar {
	for i := range rows {
		dst = append(dst, toProtoCvar(&rows[i]))
	}
	return dst
}

func (s *Server) siteTitle(ctx context.Context) string {
	row, err := s.store.FindCvar(ctx, cvar.KeySiteTitle)
	if err != nil || row == nil || row.ValueString == "" {
		return "StarApp"
	}
	return row.ValueString
}

func (s *Server) boolCvar(ctx context.Context, key string, fallback bool) bool {
	row, err := s.store.FindCvar(ctx, key)
	if err != nil || row == nil {
		return fallback
	}
	return row.ValueInt != 0
}

func (s *Server) showFooter(ctx context.Context) bool {
	fallback := true
	if s.cfg != nil {
		fallback = s.cfg.ShowFooter
	}
	return s.boolCvar(ctx, cvar.KeyShowFooter, fallback)
}

func (s *Server) redemptionApprovalDefault(ctx context.Context) bool {
	return s.boolCvar(ctx, cvar.KeyEnableRedemptionApproval, true)
}

func (s *Server) defaultAwardStars(ctx context.Context) int {
	row, err := s.store.FindCvar(ctx, cvar.KeyDefaultAwardStars)
	if err != nil || row == nil || row.ValueInt < 1 {
		return cvar.DefaultAwardStars
	}
	if row.ValueInt > cvar.MaxAwardStars {
		return cvar.MaxAwardStars
	}
	return row.ValueInt
}

func (s *Server) ListCvars(ctx context.Context, _ *connect.Request[apiv1.ListCvarsRequest]) (*connect.Response[apiv1.ListCvarsResponse], error) {
	rows, err := s.store.ListCvars(ctx)
	if err != nil {
		s.log.WithError(err).Error("list cvars")
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&apiv1.ListCvarsResponse{
		Cvars: appendProtoCvars(nil, rows),
	}), nil
}

func validateCvarUpdate(row *store.CvarRow, valueInt int32, valueString string, allowedThemes []string) (int, string, error) {
	switch row.Key {
	case cvar.KeyThemeControl:
		value := strings.TrimSpace(valueString)
		if value != cvar.ThemeControlSystem && value != cvar.ThemeControlUser {
			return 0, "", fmt.Errorf("theme_control must be system or user")
		}
		return 0, value, nil
	case cvar.KeyThemeName:
		value := strings.TrimSpace(valueString)
		if !themes.IsValidName(value, allowedThemes) {
			return 0, "", fmt.Errorf("unknown theme %q", value)
		}
		return 0, value, nil
	}

	switch row.MainType {
	case cvar.TypeString:
		return validateStringCvar(row.Key, valueString)
	case cvar.TypeTextarea:
		if len(valueString) > cvar.MaxTextareaCvarLen {
			return 0, "", fmt.Errorf("value too long")
		}
		return 0, valueString, nil
	case cvar.TypeInt:
		v := int(valueInt)
		if row.Key == cvar.KeyDefaultAwardStars && (v < 1 || v > cvar.MaxAwardStars) {
			return 0, "", fmt.Errorf("default_award_stars must be between 1 and %d", cvar.MaxAwardStars)
		}
		return v, "", nil
	case cvar.TypeBool:
		if valueInt != 0 {
			return 1, "", nil
		}
		return 0, "", nil
	default:
		return 0, "", fmt.Errorf("unsupported cvar type")
	}
}

func validateStringCvar(key, valueString string) (int, string, error) {
	if len(valueString) > cvar.MaxStringCvarLen {
		return 0, "", fmt.Errorf("value too long")
	}
	switch key {
	case cvar.KeyAppriseURL, cvar.KeyExternalBaseURL:
		return 0, strings.TrimSpace(valueString), nil
	}
	if valueString == "" {
		return 0, "", fmt.Errorf("value required")
	}
	return 0, valueString, nil
}

func (s *Server) stringCvar(ctx context.Context, key string) string {
	row, err := s.store.FindCvar(ctx, key)
	if err != nil || row == nil {
		return ""
	}
	return row.ValueString
}

func (s *Server) UpdateCvar(ctx context.Context, req *connect.Request[apiv1.UpdateCvarRequest]) (*connect.Response[apiv1.Cvar], error) {
	row, err := s.store.FindCvar(ctx, req.Msg.Key)
	if err != nil {
		s.log.WithError(err).Error("find cvar")
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if row == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("cvar not found"))
	}

	valueInt, valueString, valErr := validateCvarUpdate(row, req.Msg.ValueInt, req.Msg.ValueString, s.availableThemes(ctx))
	if valErr != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, valErr)
	}
	if updErr := s.store.UpdateCvar(ctx, row.Key, valueInt, valueString); updErr != nil {
		s.log.WithError(updErr).Error("update cvar")
		return nil, connect.NewError(connect.CodeInternal, updErr)
	}

	updated, _ := s.store.FindCvar(ctx, row.Key)
	return connect.NewResponse(toProtoCvar(updated)), nil
}
