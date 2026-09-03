package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/apprise"
	"github.com/jamesread/starapp/service/internal/config"
	"github.com/jamesread/starapp/service/internal/store"
)

func TestNotificationDeliveryHistoryRecorded(t *testing.T) {
	st := store.OpenMemory()
	t.Cleanup(func() { _ = st.Close() })
	ctx := context.Background()
	require.NoError(t, st.EnsureRBACBootstrap(ctx))
	require.NoError(t, st.SeedDomainRBAC(ctx))

	parentID, err := st.CreateUserAccount(ctx, "parent", "hash", store.UserCreatedByAdmin)
	require.NoError(t, err)
	svc := New(&config.Config{}, st, nil, logrus.New())
	parentCtx := userCtx(parentID, "parent", true)
	created, err := svc.CreateFamily(parentCtx, connect.NewRequest(&apiv1.CreateFamilyRequest{Name: "Household"}))
	require.NoError(t, err)
	parentMemberID := int(created.Msg.CallerMember.Id)

	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(func() { srv.Close() })

	err = svc.deliverApprise(ctx, &http.Client{Timeout: 2 * time.Second}, srv.URL, apprise.Payload{
		Title: "Test title",
		Body:  "Hello",
		Tag:   apprise.PersonTag(parentMemberID),
	}, int(created.Msg.Family.Id), parentMemberID, store.NotificationTypeTest)
	require.NoError(t, err)
	require.True(t, called)

	list, err := svc.ListNotificationDeliveries(parentCtx, connect.NewRequest(&apiv1.ListNotificationDeliveriesRequest{Limit: 30}))
	require.NoError(t, err)
	require.Len(t, list.Msg.Deliveries, 1)
	require.Equal(t, "Test title", list.Msg.Deliveries[0].Title)
	require.Equal(t, int32(parentMemberID), list.Msg.Deliveries[0].RecipientMemberId)
	require.True(t, list.Msg.Deliveries[0].Success)
	require.Equal(t, store.NotificationTypeTest, list.Msg.Deliveries[0].NotificationType)
}
