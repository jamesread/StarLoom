package auth

import (
	"context"
	"net/http"
	"os"

	"github.com/jamesread/armature-iam/layer"
)

func SessionCookieName() string { return "starapp-sid" }

func SecureCookiesEnabled() bool {
	return os.Getenv("STARAPP_SECURE_COOKIES") != "false"
}

func NewSessionCookie(sid string) http.Cookie {
	return http.Cookie{
		Name:     SessionCookieName(),
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

func UserFromContext(ctx context.Context) *AuthenticatedUser {
	return layer.UserFromContext(ctx)
}
