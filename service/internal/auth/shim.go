package auth

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	japauth "github.com/jamesread/httpauthshim"
	"github.com/jamesread/httpauthshim/authpublic"
	"github.com/jamesread/httpauthshim/providers/hasjwt"
	"github.com/jamesread/httpauthshim/providers/hasmtls"
	"github.com/jamesread/httpauthshim/providers/hastrustedheaders"
	"github.com/jamesread/httpauthshim/sessions"
	"github.com/sirupsen/logrus"

	"github.com/jamesread/starapp/service/internal/store"
)

const (
	providerStarAppAPIKey  = "starapp-api-key"
	providerStarAppSession = "starapp-session"
	providerJWT            = "jwt"
	cookieName             = "starapp-sid"
)

type nopSessionPersistence struct{}

func (nopSessionPersistence) Load(_, _ string, _ *sessions.SessionStorage) error { return nil }
func (nopSessionPersistence) Save(_, _ string, _ *sessions.SessionStorage) error { return nil }
func (nopSessionPersistence) RequiresFileLock() bool                             { return false }

func authConfigFromYAML(_ map[string]any) *authpublic.Config {
	cfg := &authpublic.Config{}
	if cfg.BaseDir == "" {
		cfg.BaseDir = filepath.Join(os.TempDir(), "starapp-httpauthshim-unused")
	}
	return cfg
}

func jwtConfigured(cfg *authpublic.Config) bool {
	return cfg.Jwt.CertsURL != "" || cfg.Jwt.PubKeyPath != "" || cfg.Jwt.HmacSecret != ""
}

func registerConfiguredProviders(ctx *japauth.AuthShimContext, cfg *authpublic.Config, log *logrus.Logger) {
	if jwtConfigured(cfg) {
		if cfg.Jwt.Header != "" {
			ctx.AddProvider(hasjwt.CheckUserFromJwtHeader)
			log.Info("httpauthshim: JWT header authentication enabled")
		}
		if cfg.Jwt.CookieName != "" {
			ctx.AddProvider(hasjwt.CheckUserFromJwtCookie)
			log.Info("httpauthshim: JWT cookie authentication enabled")
		}
	}
	if cfg.HttpHeader.Username != "" {
		ctx.AddProvider(hastrustedheaders.CheckUserFromHeaders)
		log.Infof("httpauthshim: trusted header authentication enabled (%s)", cfg.HttpHeader.Username)
	}
	if cfg.Mtls.Enabled {
		ctx.AddProvider(hasmtls.CheckUserFromMtls)
		log.Info("httpauthshim: mTLS authentication enabled")
	}
}

func newAuthShim(st store.Store, authYAML map[string]any, log *logrus.Logger) (*japauth.AuthShimContext, error) {
	cfg := authConfigFromYAML(authYAML)
	storage := sessions.NewSessionStorage(nopSessionPersistence{})
	ctx, err := japauth.NewAuthShimContext(cfg, storage)
	if err != nil {
		return nil, err
	}
	ctx.AddProvider(starAppAPIKeyProvider(st))
	registerConfiguredProviders(ctx, cfg, log)
	ctx.AddProvider(starAppSessionProvider(st))
	return ctx, nil
}

func extractBearerToken(req *http.Request) string {
	if req == nil {
		return ""
	}
	h := req.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
}

func starAppAPIKeyProvider(st store.Store) func(*authpublic.AuthCheckingContext) *authpublic.AuthenticatedUser {
	return func(ac *authpublic.AuthCheckingContext) *authpublic.AuthenticatedUser {
		token := extractBearerToken(ac.Request)
		if token == "" {
			return nil
		}
		user, _, err := st.GetUserByAPIKey(ac.Request.Context(), token)
		if err != nil || user == nil {
			return nil
		}
		return &authpublic.AuthenticatedUser{
			Username: user.Username,
			Provider: providerStarAppAPIKey,
		}
	}
}

func starAppSessionProvider(st store.Store) func(*authpublic.AuthCheckingContext) *authpublic.AuthenticatedUser {
	return func(ac *authpublic.AuthCheckingContext) *authpublic.AuthenticatedUser {
		if ac.Request == nil {
			return nil
		}
		c, err := ac.Request.Cookie(cookieName)
		if err != nil || c.Value == "" {
			return nil
		}
		sess, err := st.GetSessionBySID(ac.Request.Context(), c.Value)
		if err != nil || sess == nil {
			return nil
		}
		user, err := st.GetUserByID(ac.Request.Context(), sess.UserAccountID)
		if err != nil || user == nil {
			return nil
		}
		return &authpublic.AuthenticatedUser{
			Username: user.Username,
			Provider: providerStarAppSession,
			SID:      c.Value,
		}
	}
}

func SessionCookieName() string { return cookieName }

func SecureCookiesEnabled() bool {
	return os.Getenv("STARAPP_SECURE_COOKIES") != "false"
}

func NewSessionCookie(sid string) http.Cookie {
	return http.Cookie{
		Name:     cookieName,
		Value:    sid,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		Secure:   SecureCookiesEnabled(),
	}
}

func ClearSessionCookie() http.Cookie {
	c := NewSessionCookie("")
	c.MaxAge = -1
	return c
}
