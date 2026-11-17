package server

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/jamesread/armature-iam/password"
	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/auth"
	"github.com/jamesread/starapp/service/internal/rbac"
	"github.com/jamesread/starapp/service/internal/store"
)

func (s *Server) BootstrapIAM(ctx context.Context) error {
	return s.bootstrapIAM(ctx)
}

func (s *Server) bootstrapIAM(ctx context.Context) error {
	if err := s.store.EnsureRBACBootstrap(ctx); err != nil {
		return err
	}
	if err := s.store.SeedDomainRBAC(ctx); err != nil {
		return err
	}
	return s.ensureBootstrapAdmin(ctx)
}

func (s *Server) ensureBootstrapAdmin(ctx context.Context) error {
	count, err := s.store.CountUserAccounts(ctx)
	if err != nil || count > 0 {
		return err
	}
	return s.createBootstrapAdmin(ctx)
}

func (s *Server) createBootstrapAdmin(ctx context.Context) error {
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
	init := s.initShell(ctx)
	return init.SiteTitle, init
}

func (s *Server) GetStatus(ctx context.Context, req *connect.Request[apiv1.GetStatusRequest]) (*connect.Response[apiv1.GetStatusResponse], error) {
	au := auth.UserFromContext(ctx)
	_, init := s.shellFields(ctx)
	res := baseStatusResponse(init, au)
	if isLoggedInUser(au) {
		s.fillLoggedInStatus(ctx, res, au, req.Header())
	}
	return connect.NewResponse(res), nil
}

func isLoggedInUser(au *auth.AuthenticatedUser) bool {
	return au != nil && au.User != nil && au.User.ID > 0
}

func baseStatusResponse(init *apiv1.InitResponse, au *auth.AuthenticatedUser) *apiv1.GetStatusResponse {
	return &apiv1.GetStatusResponse{
		ShowFooter: init.ShowFooter, ShowNewVersions: init.ShowNewVersions,
		AvailableVersion: init.AvailableVersion, CurrentVersion: init.CurrentVersion,
		PageTitle: init.PageTitle, ShowVersionNumber: init.ShowVersionNumber,
		SiteTitle: init.SiteTitle, Features: init.Features, WebhookEvents: init.WebhookEvents,
		UsesSecureCookies: auth.SecureCookiesEnabled(),
		Username:          "<anonymous>", IsLoggedIn: isLoggedInUser(au),
	}
}

func (s *Server) fillLoggedInStatus(ctx context.Context, res *apiv1.GetStatusResponse, au *auth.AuthenticatedUser, header http.Header) {
	res.Username = au.User.Username
	res.DisplayName = s.accountDisplayName(ctx, au.User.ID, au.User.Username)
	res.IsLoggedIn = true
	if user, uerr := s.store.GetUserByID(ctx, au.User.ID); uerr == nil && user != nil {
		res.AccountCreatedAt = user.CreatedAt
	}
	fillStatusRBAC(res, au)
	s.fillStatusImpersonation(ctx, res, header)
}

func fillStatusRBAC(res *apiv1.GetStatusResponse, au *auth.AuthenticatedUser) {
	if au.RBAC == nil {
		return
	}
	res.RbacIsSuperuser = au.RBAC.IsSuperuser
	for p := range au.RBAC.Permissions {
		res.RbacPermissions = append(res.RbacPermissions, p)
	}
	sort.Strings(res.RbacPermissions)
}

func (s *Server) fillStatusImpersonation(ctx context.Context, res *apiv1.GetStatusResponse, header http.Header) {
	sid := sessionIDFromHeader(header)
	if sid == "" {
		return
	}
	impersonatorID, ok := s.sessionImpersonatorID(ctx, sid)
	if !ok {
		return
	}
	res.IsImpersonating = true
	if imp, _ := s.store.GetUserByID(ctx, impersonatorID); imp != nil {
		res.ImpersonatorUsername = imp.Username
	}
}

func (s *Server) sessionImpersonatorID(ctx context.Context, sid string) (int, bool) {
	sess, err := s.store.GetSessionBySID(ctx, sid)
	if err != nil || sess == nil || sess.ImpersonatorUserID == nil {
		return 0, false
	}
	return *sess.ImpersonatorUserID, true
}

func sessionIDFromHeader(h http.Header) string {
	prefix := auth.SessionCookieName() + "="
	for part := range strings.SplitSeq(h.Get("Cookie"), ";") {
		part = strings.TrimSpace(part)
		if val, ok := strings.CutPrefix(part, prefix); ok {
			return val
		}
	}
	return ""
}

func (s *Server) LoginWithUsernameAndPassword(ctx context.Context, req *connect.Request[apiv1.LoginWithUsernameAndPasswordRequest]) (*connect.Response[apiv1.LoginWithUsernameAndPasswordResponse], error) {
	user, err := s.store.GetUserByUsername(ctx, strings.TrimSpace(req.Msg.Username))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	res := failedLoginResponse()
	if !passwordMatches(user, req.Msg.Password) {
		return res, nil
	}
	return s.completeLogin(ctx, res, user)
}

func failedLoginResponse() *connect.Response[apiv1.LoginWithUsernameAndPasswordResponse] {
	return connect.NewResponse(&apiv1.LoginWithUsernameAndPasswordResponse{
		StandardResponse: &apiv1.StandardResponse{Success: false, Message: "Invalid username or password"},
	})
}

func passwordMatches(user *store.UserAccountRow, plain string) bool {
	if user == nil || user.PasswordHash == "" {
		return false
	}
	match, err := password.Verify(user.PasswordHash, plain)
	return err == nil && match
}

func (s *Server) completeLogin(ctx context.Context, res *connect.Response[apiv1.LoginWithUsernameAndPasswordResponse], user *store.UserAccountRow) (*connect.Response[apiv1.LoginWithUsernameAndPasswordResponse], error) {
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
	au, err := s.requireWritableUser(ctx)
	if err != nil {
		return nil, err
	}
	if len(req.Msg.NewPassword) < 8 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("new password must be at least 8 characters"))
	}
	if err := s.changeUserPassword(ctx, au.User.ID, req.Msg.CurrentPassword, req.Msg.NewPassword); err != nil {
		return nil, err
	}
	return connect.NewResponse(&apiv1.ChangePasswordResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Password changed"},
	}), nil
}

func (s *Server) requireWritableUser(ctx context.Context) (*auth.AuthenticatedUser, error) {
	au := auth.UserFromContext(ctx)
	if au == nil || au.User == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("authentication required"))
	}
	if au.ReadOnly {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("read-only API key"))
	}
	return au, nil
}

func (s *Server) changeUserPassword(ctx context.Context, userID int, currentPassword, newPassword string) error {
	user, err := s.passwordUser(ctx, userID)
	if err != nil {
		return err
	}
	if vErr := verifyCurrentPassword(user.PasswordHash, currentPassword); vErr != nil {
		return vErr
	}
	hash, err := password.Hash(newPassword)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if err := s.store.UpdateUserPassword(ctx, user.ID, hash); err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}

func (s *Server) passwordUser(ctx context.Context, userID int) (*store.UserAccountRow, error) {
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if user.PasswordHash == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user has no password set"))
	}
	return user, nil
}

func verifyCurrentPassword(hash, currentPassword string) error {
	match, err := password.Verify(hash, currentPassword)
	if err != nil || !match {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid current password"))
	}
	return nil
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

func (s *Server) accountDisplayName(ctx context.Context, userID int, username string) string {
	member, err := s.store.GetMemberByAccountID(ctx, userID)
	if err == nil && member != nil {
		if name := strings.TrimSpace(member.DisplayName); name != "" {
			return name
		}
	}
	return username
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
