package auth

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/jamesread/armature-iam/layer"
	"github.com/jamesread/starapp/service/gen/starapp/api/v1/apiv1connect"
	"github.com/jamesread/starapp/service/internal/store"
)

type Layer = layer.Layer
type AuthenticatedUser = layer.AuthenticatedUser

func NewLayer(st store.Store, authYAML map[string]any, log *logrus.Logger) *layer.Layer {
	if log == nil {
		log = logrus.New()
	}
	l, err := layer.New(st, layer.Config{
		CookieName: SessionCookieName(),
		AuthYAML:   authYAML,
		Logger:     log,
		AllowUnauthenticated: []string{
			apiv1connect.StarAppServiceLoginWithUsernameAndPasswordProcedure,
			apiv1connect.StarAppServiceGetStatusProcedure,
		},
		RequiredPermission: RequiredPermission,
		DevDisableAuth:     os.Getenv("STARAPP_DEV_DISABLE_AUTH") == "true",
		APIKeyPrefix:       "sa_",
		SecureCookies:      SecureCookiesEnabled(),
	})
	if err != nil {
		log.Fatalf("armature-iam init: %v", err)
	}
	return l
}
