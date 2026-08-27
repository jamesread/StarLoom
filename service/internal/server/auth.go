package server

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/auth"
	"github.com/jamesread/starapp/service/internal/buildinfo"
	"github.com/jamesread/starapp/service/internal/password"
	"github.com/jamesread/starapp/service/internal/rbac"
	"github.com/jamesread/starapp/service/internal/store"
	"github.com/jamesread/starapp/service/internal/webhook"
)

func (s *Server) BootstrapIAM(ctx context.Context) error {
	return s.bootstrapIAM(ctx)
}

func (s *Server) bootstrapIAM(ctx context.Context) error {
	if err := s.store.EnsureRBACBootstrap(ctx); err != nil {
		return err
	}
	count, err := s.store.CountUserAccounts(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hash, err := password.Hash("admin")
	if err != nil {
		return err
	}
	id, err := s.store.CreateUserAccount(ctx, "admin", hash, store.UserCreatedByAdmin)
	if err != nil {
		return err
	}
	if err := s.store.EnsureUserInEveryoneGroup(ctx, id); err != nil {
		return err
	}
	if err := s.store.EnsureUserInGroup(ctx, id, rbac.GroupParents); err != nil {
		return err
	}
	return s.store.EnsureRBACBootstrap(ctx)
}

func (s *Server) shellFields(ctx context.Context) (title string, resp *apiv1.InitResponse) {
	title = s.siteTitle(ctx)
	return title, &apiv1.InitResponse{
		ShowFooter:        s.showFooter(ctx),
		ShowNewVersions:   false,
		AvailableVersion:  "",
		CurrentVersion:    buildinfo.Version,
		PageTitle:         title,
		ShowVersionNumber: true,
		SiteTitle:         title,
		Features: &apiv1.Features{
			RedemptionApprovalDefault: s.redemptionApprovalDefault(ctx),
		},
		WebhookEvents: append([]string(nil), webhook.SupportedEvents...),
	}
}

func (s *Server) GetStatus(ctx context.Context, req *connect.Request[apiv1.GetStatusRequest]) (*connect.Response[apiv1.GetStatusResponse], error) {
	au := auth.UserFromContext(ctx)
	_, init := s.shellFields(ctx)

	res := &apiv1.GetStatusResponse{
		ShowFooter:        init.ShowFooter,
		ShowNewVersions:   init.ShowNewVersions,
		AvailableVersion:  init.AvailableVersion,
		CurrentVersion:    init.CurrentVersion,
		PageTitle:         init.PageTitle,
		ShowVersionNumber: init.ShowVersionNumber,
		SiteTitle:         init.SiteTitle,
		Features:          init.Features,
		WebhookEvents:     init.WebhookEvents,
		UsesSecureCookies: auth.SecureCookiesEnabled(),
		Username:          "<anonymous>",
		IsLoggedIn:        au != nil && au.User != nil && au.User.ID > 0,
	}

	if au != nil && au.User != nil && au.User.ID > 0 {
		res.Username = au.User.Username
		res.IsLoggedIn = true
		if user, uerr := s.store.GetUserByID(ctx, au.User.ID); uerr == nil && user != nil {
			res.AccountCreatedAt = user.CreatedAt
		}
		if au.RBAC != nil {
			res.RbacIsSuperuser = au.RBAC.IsSuperuser
			for p := range au.RBAC.Permissions {
				res.RbacPermissions = append(res.RbacPermissions, p)
			}
			sort.Strings(res.RbacPermissions)
		}
		if sid := sessionIDFromHeader(req.Header()); sid != "" {
			if sess, err := s.store.GetSessionBySID(ctx, sid); err == nil && sess != nil && sess.ImpersonatorUserID != nil {
				res.IsImpersonating = true
				if imp, _ := s.store.GetUserByID(ctx, *sess.ImpersonatorUserID); imp != nil {
					res.ImpersonatorUsername = imp.Username
				}
			}
		}
	}

	return connect.NewResponse(res), nil
}

func sessionIDFromHeader(h http.Header) string {
	for _, part := range strings.Split(h.Get("Cookie"), ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, auth.SessionCookieName()+"=") {
			return strings.TrimPrefix(part, auth.SessionCookieName()+"=")
		}
	}
	return ""
}

func (s *Server) LoginWithUsernameAndPassword(ctx context.Context, req *connect.Request[apiv1.LoginWithUsernameAndPasswordRequest]) (*connect.Response[apiv1.LoginWithUsernameAndPasswordResponse], error) {
	username := strings.TrimSpace(req.Msg.Username)
	user, err := s.store.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	res := connect.NewResponse(&apiv1.LoginWithUsernameAndPasswordResponse{
		StandardResponse: &apiv1.StandardResponse{Success: false, Message: "Invalid username or password"},
	})
	if user == nil || user.PasswordHash == "" {
		return res, nil
	}
	match, err := password.Verify(user.PasswordHash, req.Msg.Password)
	if err != nil || !match {
		return res, nil
	}

	sid := uuid.NewString()
	if err := s.store.CreateSession(ctx, sid, user.ID, nil); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	res.Msg.StandardResponse = &apiv1.StandardResponse{Success: true, Message: "Login successful"}
	res.Msg.Username = user.Username
	c := auth.NewSessionCookie(sid)
	res.Header().Add("Set-Cookie", c.String())
	return res, nil
}

func (s *Server) Logout(ctx context.Context, req *connect.Request[apiv1.LogoutRequest]) (*connect.Response[apiv1.LogoutResponse], error) {
	if sid := sessionIDFromHeader(req.Header()); sid != "" {
		_ = s.store.DeleteSession(ctx, sid)
	}
	res := connect.NewResponse(&apiv1.LogoutResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Logged out"},
	})
	c := auth.ClearSessionCookie()
	res.Header().Add("Set-Cookie", c.String())
	return res, nil
}

func (s *Server) ChangePassword(ctx context.Context, req *connect.Request[apiv1.ChangePasswordRequest]) (*connect.Response[apiv1.ChangePasswordResponse], error) {
	au := auth.UserFromContext(ctx)
	if au == nil || au.User == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("authentication required"))
	}
	if au.ReadOnly {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("read-only API key"))
	}
	if len(req.Msg.NewPassword) < 8 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("new password must be at least 8 characters"))
	}
	user, err := s.store.GetUserByID(ctx, au.User.ID)
	if err != nil || user == nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if user.PasswordHash == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user has no password set"))
	}
	match, err := password.Verify(user.PasswordHash, req.Msg.CurrentPassword)
	if err != nil || !match {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid current password"))
	}
	hash, err := password.Hash(req.Msg.NewPassword)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if err := s.store.UpdateUserPassword(ctx, user.ID, hash); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&apiv1.ChangePasswordResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Password changed"},
	}), nil
}

func toProtoUser(u *store.UserAccountRow) *apiv1.UserAccount {
	if u == nil {
		return nil
	}
	return &apiv1.UserAccount{
		Id:        int32(u.ID),
		Username:  u.Username,
		CreatedAt: u.CreatedAt,
		CreatedBy: u.CreatedBy,
	}
}

func (s *Server) requireAuth(ctx context.Context) (*auth.AuthenticatedUser, error) {
	au := auth.UserFromContext(ctx)
	if au == nil || au.User == nil || au.User.ID == 0 {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("authentication required"))
	}
	return au, nil
}

func (s *Server) requireWrite(ctx context.Context) (*auth.AuthenticatedUser, error) {
	au, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if au.ReadOnly {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("read-only API key"))
	}
	return au, nil
}
