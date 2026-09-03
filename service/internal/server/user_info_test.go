package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/apprise"
	"github.com/jamesread/starapp/service/internal/config"
	"github.com/jamesread/starapp/service/internal/store"
)

func TestGetUserIncludesLinkedMember(t *testing.T) {
	st := store.OpenMemory()
	t.Cleanup(func() { _ = st.Close() })
	ctx := context.Background()
	require.NoError(t, st.EnsureRBACBootstrap(ctx))
	require.NoError(t, st.SeedDomainRBAC(ctx))

	parentID, err := st.CreateUserAccount(ctx, "parent", "hash", store.UserCreatedByAdmin)
	require.NoError(t, err)
	childID, err := st.CreateUserAccount(ctx, "child", "hash", store.UserCreatedByAdmin)
	require.NoError(t, err)

	svc := New(&config.Config{}, st, nil, logrus.New())
	parentCtx := userCtx(parentID, "parent", true)
	created, err := svc.CreateFamily(parentCtx, connect.NewRequest(&apiv1.CreateFamilyRequest{Name: "Household"}))
	require.NoError(t, err)
	childMemberID, err := st.CreateMember(ctx, int(created.Msg.Family.Id), "Sam", store.MemberRoleChild, &childID, "")
	require.NoError(t, err)

	got, err := svc.GetUser(parentCtx, connect.NewRequest(&apiv1.GetUserRequest{UserId: int32(childID)}))
	require.NoError(t, err)
	require.Equal(t, "child", got.Msg.User.Username)
	require.NotNil(t, got.Msg.LinkedMember)
	require.Equal(t, int32(childMemberID), got.Msg.LinkedMember.Id)
	require.Equal(t, "Sam", got.Msg.LinkedMember.DisplayName)
}

func TestGetUserIncludesUserGroups(t *testing.T) {
	st := store.OpenMemory()
	t.Cleanup(func() { _ = st.Close() })
	ctx := context.Background()
	require.NoError(t, st.EnsureRBACBootstrap(ctx))
	require.NoError(t, st.SeedDomainRBAC(ctx))

	parentID, err := st.CreateUserAccount(ctx, "parent", "hash", store.UserCreatedByAdmin)
	require.NoError(t, err)
	require.NoError(t, st.EnsureUserInEveryoneGroup(ctx, parentID))
	require.NoError(t, st.EnsureUserInGroup(ctx, parentID, "Parents"))

	svc := New(&config.Config{}, st, nil, logrus.New())
	parentCtx := userCtx(parentID, "parent", true)

	got, err := svc.GetUser(parentCtx, connect.NewRequest(&apiv1.GetUserRequest{UserId: int32(parentID)}))
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(got.Msg.UserGroups), 2)
	names := make([]string, 0, len(got.Msg.UserGroups))
	for _, g := range got.Msg.UserGroups {
		names = append(names, g.Name)
	}
	require.Contains(t, names, "Everyone")
	require.Contains(t, names, "Parents")
}

func TestSendUserTestNotification(t *testing.T) {
	st := store.OpenMemory()
	t.Cleanup(func() { _ = st.Close() })
	ctx := context.Background()
	require.NoError(t, st.EnsureRBACBootstrap(ctx))
	require.NoError(t, st.SeedDomainRBAC(ctx))

	parentID, err := st.CreateUserAccount(ctx, "parent", "hash", store.UserCreatedByAdmin)
	require.NoError(t, err)
	childID, err := st.CreateUserAccount(ctx, "child", "hash", store.UserCreatedByAdmin)
	require.NoError(t, err)

	svc := New(&config.Config{}, st, nil, logrus.New())
	parentCtx := userCtx(parentID, "parent", true)
	created, err := svc.CreateFamily(parentCtx, connect.NewRequest(&apiv1.CreateFamilyRequest{Name: "Household"}))
	require.NoError(t, err)
	childMemberID, err := st.CreateMember(ctx, int(created.Msg.Family.Id), "Sam", store.MemberRoleChild, &childID, "")
	require.NoError(t, err)

	require.NoError(t, st.InsertCvarIfMissing(ctx, store.CvarRow{
		Key: "apprise_url", MainType: "string", ValueString: "http://example.test/notify",
	}))

	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(func() { srv.Close() })
	require.NoError(t, st.UpdateCvar(ctx, "apprise_url", 0, srv.URL))

	sent, err := svc.SendUserTestNotification(parentCtx, connect.NewRequest(&apiv1.SendUserTestNotificationRequest{
		UserId: int32(childID),
	}))
	require.NoError(t, err)
	require.True(t, sent.Msg.StandardResponse.Success)
	require.Equal(t, apprise.PersonTag(childMemberID), sent.Msg.Tag)
	require.True(t, called)

	list, err := st.ListNotificationDeliveries(ctx, 10)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, childMemberID, list[0].RecipientMemberID)
	require.True(t, list[0].Success)
}

func TestSendUserTestNotificationRequiresLinkedMember(t *testing.T) {
	st := store.OpenMemory()
	t.Cleanup(func() { _ = st.Close() })
	ctx := context.Background()
	require.NoError(t, st.EnsureRBACBootstrap(ctx))
	require.NoError(t, st.SeedDomainRBAC(ctx))

	parentID, err := st.CreateUserAccount(ctx, "parent", "hash", store.UserCreatedByAdmin)
	require.NoError(t, err)
	orphanID, err := st.CreateUserAccount(ctx, "orphan", "hash", store.UserCreatedByAdmin)
	require.NoError(t, err)

	svc := New(&config.Config{}, st, nil, logrus.New())
	parentCtx := userCtx(parentID, "parent", true)
	_, err = svc.SendUserTestNotification(parentCtx, connect.NewRequest(&apiv1.SendUserTestNotificationRequest{
		UserId: int32(orphanID),
	}))
	require.Error(t, err)
	require.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
}
