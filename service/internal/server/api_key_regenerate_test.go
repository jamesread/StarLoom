package server

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/config"
	"github.com/jamesread/starapp/service/internal/store"
)

func TestRegenerateApiKeyRevokesTheOldSecret(t *testing.T) {
	st := store.OpenMemory()
	ctx := superuserCtx(context.Background())
	_, err := st.CreateUserAccount(ctx, "admin", "", store.UserCreatedByAdmin)
	require.NoError(t, err)
	svc := New(&config.Config{}, st, nil, logrus.New())

	created, err := svc.CreateApiKey(ctx, connect.NewRequest(&apiv1.CreateApiKeyRequest{
		Name:     "Kitchen display",
		ReadOnly: true,
	}))
	require.NoError(t, err)

	fresh, err := svc.RegenerateApiKey(ctx, connect.NewRequest(&apiv1.RegenerateApiKeyRequest{
		Id: created.Msg.Key.Id,
	}))
	require.NoError(t, err)
	require.NotEqual(t, created.Msg.Secret, fresh.Msg.Secret)
	require.Equal(t, "Kitchen display", fresh.Msg.Key.Name)
	require.True(t, fresh.Msg.Key.ReadOnly)

	user, _, err := svc.store.GetUserByAPIKey(ctx, created.Msg.Secret)
	require.NoError(t, err)
	require.Nil(t, user)

	user, readOnly, err := svc.store.GetUserByAPIKey(ctx, fresh.Msg.Secret)
	require.NoError(t, err)
	require.NotNil(t, user)
	require.True(t, readOnly)

	keys, err := svc.store.ListAPIKeysForUser(ctx, 1)
	require.NoError(t, err)
	require.Len(t, keys, 1)
}

func TestRegenerateApiKeyRejectsUnknownID(t *testing.T) {
	st := store.OpenMemory()
	ctx := superuserCtx(context.Background())
	_, err := st.CreateUserAccount(ctx, "admin", "", store.UserCreatedByAdmin)
	require.NoError(t, err)
	svc := New(&config.Config{}, st, nil, logrus.New())

	_, err = svc.RegenerateApiKey(ctx, connect.NewRequest(&apiv1.RegenerateApiKeyRequest{Id: 999}))
	require.Error(t, err)
	require.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}
