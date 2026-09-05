package server

import (
	"context"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/config"
	"github.com/jamesread/starapp/service/internal/store"
)

func bearerTokenServer(t *testing.T) (*Server, context.Context) {
	t.Helper()
	st := store.OpenMemory()
	ctx := superuserCtx(context.Background())
	_, err := st.CreateUserAccount(ctx, "admin", "", store.UserCreatedByAdmin)
	require.NoError(t, err)
	return New(&config.Config{}, st, nil, logrus.New()), ctx
}

func TestGetBearerTokenIsDurable(t *testing.T) {
	svc, ctx := bearerTokenServer(t)

	first, err := svc.GetBearerToken(ctx, connect.NewRequest(&apiv1.GetBearerTokenRequest{}))
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(first.Msg.Token, "sa_"))

	second, err := svc.GetBearerToken(ctx, connect.NewRequest(&apiv1.GetBearerTokenRequest{}))
	require.NoError(t, err)
	require.Equal(t, first.Msg.Token, second.Msg.Token)
}

func TestRegenerateBearerTokenRevokesTheOldToken(t *testing.T) {
	svc, ctx := bearerTokenServer(t)

	old, err := svc.GetBearerToken(ctx, connect.NewRequest(&apiv1.GetBearerTokenRequest{}))
	require.NoError(t, err)

	fresh, err := svc.RegenerateBearerToken(ctx, connect.NewRequest(&apiv1.RegenerateBearerTokenRequest{}))
	require.NoError(t, err)
	require.NotEqual(t, old.Msg.Token, fresh.Msg.Token)

	user, _, err := svc.store.GetUserByAPIKey(ctx, old.Msg.Token)
	require.NoError(t, err)
	require.Nil(t, user)

	user, readOnly, err := svc.store.GetUserByAPIKey(ctx, fresh.Msg.Token)
	require.NoError(t, err)
	require.NotNil(t, user)
	require.True(t, readOnly)

	keys, err := svc.store.ListAPIKeysForUser(ctx, 1)
	require.NoError(t, err)
	require.Len(t, keys, 1)
}
