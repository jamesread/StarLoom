package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"connectrpc.com/authn"
	japauth "github.com/jamesread/httpauthshim"
	"github.com/jamesread/httpauthshim/authpublic"
	"github.com/sirupsen/logrus"

	apiv1connect "github.com/jamesread/starapp/service/gen/starapp/api/v1/apiv1connect"
	"github.com/jamesread/starapp/service/internal/rbac"
	"github.com/jamesread/starapp/service/internal/store"
)

var allowList = map[string]bool{
	apiv1connect.StarAppServiceLoginWithUsernameAndPasswordProcedure: true,
	apiv1connect.StarAppServiceGetStatusProcedure:                    true,
}

type Layer struct {
	Store     store.Store
	shim      *japauth.AuthShimContext
	devNoAuth bool
	log       *logrus.Logger
}

func NewLayer(st store.Store, authYAML map[string]any, log *logrus.Logger) *Layer {
	if log == nil {
		log = logrus.New()
	}
	if os.Getenv("STARAPP_DEV_DISABLE_AUTH") == "true" {
		log.Warn("STARAPP_DEV_DISABLE_AUTH is set: all API requests run as anonymous superuser")
		return &Layer{Store: st, devNoAuth: true, log: log}
	}
	shim, err := newAuthShim(st, authYAML, log)
	if err != nil {
		log.Fatalf("httpauthshim init: %v", err)
	}
	return &Layer{Store: st, shim: shim, log: log}
}

func (l *Layer) finishWithRBAC(ctx context.Context, au *AuthenticatedUser, procedureName string) (any, error) {
	rb, err := l.Store.LoadEffectiveRBAC(ctx, au.User.ID)
	if err != nil {
		l.log.WithError(err).Error("LoadEffectiveRBAC")
		return nil, authn.Errorf("Authentication Required")
	}
	au.RBAC = rb

	if allowList[procedureName] {
		return au, nil
	}

	req := RequiredPermission(procedureName)
	if req != "" && !au.HasPermission(req) {
		l.log.WithFields(logrus.Fields{
			"user":      au.User.Username,
			"procedure": procedureName,
			"perm":      req,
		}).Warn("RBAC denied")
		return nil, authn.Errorf("Forbidden")
	}
	return au, nil
}

func (l *Layer) Handle(ctx context.Context, req *http.Request) (any, error) {
	procedureName, _ := authn.InferProcedure(req.URL)

	if l.devNoAuth {
		au := &AuthenticatedUser{
			User: &store.UserAccountRow{ID: 0, Username: "anonymous"},
			RBAC: &rbac.EffectiveRBAC{IsSuperuser: true, Permissions: map[string]bool{}},
		}
		if allowList[procedureName] {
			return au, nil
		}
		if perm := RequiredPermission(procedureName); perm != "" && !au.HasPermission(perm) {
			return nil, authn.Errorf("Forbidden")
		}
		return au, nil
	}

	if token, ok := authn.BearerToken(req); ok {
		user, readOnly, err := l.Store.GetUserByAPIKey(ctx, token)
		if err == nil && user != nil {
			au := &AuthenticatedUser{User: user, ReadOnly: readOnly}
			return l.finishWithRBAC(ctx, au, procedureName)
		}
	}

	shimUser, err := l.shim.AuthFromHttpReqWithError(req)
	if err != nil {
		return nil, authn.Errorf("Authentication Required")
	}

	if shimUser.IsGuest() {
		if allowList[procedureName] {
			return nil, nil
		}
		return nil, authn.Errorf("Authentication Required")
	}

	dbUser, err := l.resolveDBUser(ctx, shimUser)
	if err != nil || dbUser == nil {
		l.log.WithField("username", shimUser.Username).Warn("session user not in database")
		return nil, authn.Errorf("Authentication Required")
	}

	au := &AuthenticatedUser{User: dbUser}
	if shimUser.Provider == providerStarAppAPIKey {
		_, readOnly, _ := l.Store.GetUserByAPIKey(ctx, extractBearerToken(req))
		au.ReadOnly = readOnly
	}
	return l.finishWithRBAC(ctx, au, procedureName)
}

func (l *Layer) resolveDBUser(ctx context.Context, shimUser *authpublic.AuthenticatedUser) (*store.UserAccountRow, error) {
	if shimUser == nil || shimUser.Username == "" {
		return nil, nil
	}
	user, err := l.Store.GetUserByUsername(ctx, shimUser.Username)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return user, nil
	}
	if shimUser.Provider != providerJWT {
		return nil, nil
	}
	id, err := l.Store.CreateUserAccount(ctx, shimUser.Username, "", store.UserCreatedBySSO)
	if err != nil {
		if existing, _ := l.Store.GetUserByUsername(ctx, shimUser.Username); existing != nil {
			return existing, nil
		}
		return nil, err
	}
	if err := l.Store.EnsureUserInEveryoneGroup(ctx, id); err != nil {
		l.log.WithError(err).Warn("EnsureUserInEveryoneGroup for SSO user")
	}
	return l.Store.GetUserByID(ctx, id)
}

func (l *Layer) WrapHandler(in http.Handler) http.Handler {
	return authn.NewMiddleware(l.Handle).Wrap(in)
}

func (l *Layer) WrapMCPHandler(in http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		au, status, err := l.authenticateMCP(r)
		if err != nil {
			writeMCPAuthError(w, status, err.Error())
			return
		}
		ctx := authn.SetInfo(r.Context(), au)
		in.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeMCPAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (l *Layer) authenticateMCP(r *http.Request) (*AuthenticatedUser, int, error) {
	ctx := r.Context()
	if l.devNoAuth {
		return &AuthenticatedUser{
			User: &store.UserAccountRow{Username: "anonymous"},
			RBAC: &rbac.EffectiveRBAC{IsSuperuser: true, Permissions: map[string]bool{}},
		}, http.StatusOK, nil
	}
	if extractBearerToken(r) == "" {
		return nil, http.StatusUnauthorized, errMCPAuth
	}
	shimUser, err := l.shim.AuthFromHttpReqWithError(r)
	if err != nil {
		return nil, http.StatusUnauthorized, errMCPAuth
	}
	if shimUser.IsGuest() || shimUser.Provider != providerStarAppAPIKey {
		return nil, http.StatusUnauthorized, errMCPAuth
	}
	user, readOnly, err := l.Store.GetUserByAPIKey(ctx, extractBearerToken(r))
	if err != nil || user == nil {
		return nil, http.StatusUnauthorized, errMCPAuth
	}
	au := &AuthenticatedUser{User: user, ReadOnly: readOnly}
	info, rbacErr := l.finishWithRBAC(ctx, au, "")
	if rbacErr != nil {
		return nil, http.StatusForbidden, errMCPForbidden
	}
	return info.(*AuthenticatedUser), http.StatusOK, nil
}

type mcpAuthErr struct{ msg string }

func (e *mcpAuthErr) Error() string { return e.msg }

var (
	errMCPAuth      = &mcpAuthErr{msg: "Authorization required: Bearer API key"}
	errMCPForbidden = &mcpAuthErr{msg: "Forbidden"}
)

func UserFromContext(ctx context.Context) *AuthenticatedUser {
	info := authn.GetInfo(ctx)
	if info == nil {
		return nil
	}
	au, _ := info.(*AuthenticatedUser)
	return au
}
