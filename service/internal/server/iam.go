package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/jamesread/armature-iam/password"
	iamstore "github.com/jamesread/armature-iam/store"
	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/auth"
	"github.com/jamesread/starapp/service/internal/store"
)

func toProtoRbacPermission(p *store.RBACPermissionRow) *apiv1.RbacPermission {
	if p == nil {
		return nil
	}
	return &apiv1.RbacPermission{
		Id:          int32(p.ID),
		Name:        p.Name,
		Description: p.Description,
	}
}

func toProtoRbacRole(r *store.RBACRoleRow) *apiv1.RbacRole {
	if r == nil {
		return nil
	}
	return &apiv1.RbacRole{
		Id:            int32(r.ID),
		Name:          r.Name,
		Description:   r.Description,
		PermissionIds: int32Slice(r.PermissionIDs),
		GroupCount:    int32(r.GroupCount),
		UserCount:     int32(r.UserCount),
	}
}

func toProtoUserGroup(g *store.UserGroupRow) *apiv1.UserGroup {
	if g == nil {
		return nil
	}
	return &apiv1.UserGroup{
		Id:          int32(g.ID),
		Name:        g.Name,
		MemberCount: int32(g.MemberCount),
	}
}

func toProtoApiKey(k *store.APIKeyRow) *apiv1.ApiKey {
	if k == nil {
		return nil
	}
	return &apiv1.ApiKey{
		Id:         int32(k.ID),
		Name:       k.Name,
		CreatedAt:  k.CreatedAt,
		LastUsedAt: k.LastUsedAt,
		ReadOnly:   k.ReadOnly,
	}
}

func toProtoMyPermissionAuditRow(r *store.MyPermissionAuditRow) *apiv1.MyPermissionAuditRow {
	if r == nil {
		return nil
	}
	return &apiv1.MyPermissionAuditRow{
		Permission:     r.Permission,
		Granted:        r.Granted,
		GrantingGroups: append([]string(nil), r.GrantingGroups...),
	}
}

func int32Slice(ids []int) []int32 {
	out := make([]int32, len(ids))
	for i, id := range ids {
		out[i] = int32(id)
	}
	return out
}

func intSlice(ids []int32) []int {
	out := make([]int, len(ids))
	for i, id := range ids {
		out[i] = int(id)
	}
	return out
}

func mapStoreError(err error) *connect.Error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return connect.NewError(connect.CodeNotFound, err)
	}
	if errors.Is(err, iamstore.ErrNoSuperuser) ||
		errors.Is(err, iamstore.ErrSystemRole) ||
		errors.Is(err, iamstore.ErrSystemGroup) ||
		errors.Is(err, iamstore.ErrRenameSystemRole) {
		return connect.NewError(connect.CodeFailedPrecondition, err)
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "already exists"):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case strings.Contains(msg, "cannot delete system"),
		strings.Contains(msg, "cannot modify system"),
		strings.Contains(msg, "cannot rename system"),
		strings.Contains(msg, "cannot set permissions for system"),
		strings.Contains(msg, "refusing to leave the system without a superuser"):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}

func (s *Server) impersonatorUserID(ctx context.Context, header http.Header, au *auth.AuthenticatedUser) (int, error) {
	if sid := sessionIDFromHeader(header); sid != "" {
		sess, err := s.store.GetSessionBySID(ctx, sid)
		if err != nil {
			return 0, connect.NewError(connect.CodeInternal, err)
		}
		if sess != nil && sess.ImpersonatorUserID != nil {
			return *sess.ImpersonatorUserID, nil
		}
	}
	return au.User.ID, nil
}

func (s *Server) GetUsers(ctx context.Context, _ *connect.Request[apiv1.GetUsersRequest]) (*connect.Response[apiv1.GetUsersResponse], error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	rows, err := s.store.ListUserAccounts(ctx)
	if err != nil {
		s.log.WithError(err).Error("list user accounts")
		return nil, mapStoreError(err)
	}
	out := &apiv1.GetUsersResponse{}
	for i := range rows {
		out.Users = append(out.Users, toProtoUser(&rows[i]))
	}
	return connect.NewResponse(out), nil
}

func (s *Server) GetUser(ctx context.Context, req *connect.Request[apiv1.GetUserRequest]) (*connect.Response[apiv1.GetUserResponse], error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	out, err := s.getUserResponse(ctx, int(req.Msg.UserId))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			return nil, err
		}
		s.log.WithError(err).Error("get user")
		return nil, err
	}
	return connect.NewResponse(out), nil
}

func (s *Server) getUserResponse(ctx context.Context, userID int) (*apiv1.GetUserResponse, error) {
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return nil, mapStoreError(err)
	}
	if user == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found"))
	}
	out := &apiv1.GetUserResponse{User: toProtoUser(user)}
	member, err := s.store.GetMemberByAccountID(ctx, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if member != nil {
		out.LinkedMember = toProtoMember(member)
	}
	groupIDs, err := s.store.ListUserGroupIDsForUser(ctx, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	for _, groupID := range groupIDs {
		group, err := s.store.GetUserGroupByID(ctx, groupID)
		if err != nil || group == nil {
			continue
		}
		out.UserGroups = append(out.UserGroups, toProtoUserGroup(group))
	}
	sort.Slice(out.UserGroups, func(i, j int) bool {
		return out.UserGroups[i].Name < out.UserGroups[j].Name
	})
	return out, nil
}

func (s *Server) CreateUser(ctx context.Context, req *connect.Request[apiv1.CreateUserRequest]) (*connect.Response[apiv1.CreateUserResponse], error) {
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	username := strings.TrimSpace(req.Msg.Username)
	if username == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("username required"))
	}
	pwd := req.Msg.Password
	if pwd != "" && len(pwd) < 8 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("password must be at least 8 characters"))
	}
	hash := ""
	if pwd != "" {
		var err error
		hash, err = password.Hash(pwd)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}
	id, err := s.store.CreateUserAccount(ctx, username, hash, store.UserCreatedByAdmin)
	if err != nil {
		s.log.WithError(err).Error("create user account")
		return nil, mapStoreError(err)
	}
	if err := s.store.EnsureUserInEveryoneGroup(ctx, id); err != nil {
		s.log.WithError(err).Error("ensure user in everyone group")
		return nil, mapStoreError(err)
	}
	if err := s.store.EnsureRBACBootstrap(ctx); err != nil {
		s.log.WithError(err).Error("ensure rbac bootstrap")
		return nil, mapStoreError(err)
	}
	user, err := s.store.GetUserByID(ctx, id)
	if err != nil {
		s.log.WithError(err).Error("get created user")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.CreateUserResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "User created"},
		User:             toProtoUser(user),
	}), nil
}

func (s *Server) DeleteUser(ctx context.Context, req *connect.Request[apiv1.DeleteUserRequest]) (*connect.Response[apiv1.DeleteUserResponse], error) {
	au, err := s.requireWrite(ctx)
	if err != nil {
		return nil, err
	}
	userID := int(req.Msg.UserId)
	if userID == au.User.ID {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("cannot delete your own account"))
	}
	if err := s.store.DeleteUserAccount(ctx, userID); err != nil {
		s.log.WithError(err).Error("delete user account")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.DeleteUserResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "User deleted"},
	}), nil
}

func (s *Server) ResetUserPassword(ctx context.Context, req *connect.Request[apiv1.ResetUserPasswordRequest]) (*connect.Response[apiv1.ResetUserPasswordResponse], error) {
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if len(req.Msg.NewPassword) < 8 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("new password must be at least 8 characters"))
	}
	hash, err := password.Hash(req.Msg.NewPassword)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if err := s.store.UpdateUserPassword(ctx, int(req.Msg.UserId), hash); err != nil {
		s.log.WithError(err).Error("reset user password")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.ResetUserPasswordResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Password reset"},
	}), nil
}

func (s *Server) ListRbacPermissions(ctx context.Context, _ *connect.Request[apiv1.ListRbacPermissionsRequest]) (*connect.Response[apiv1.ListRbacPermissionsResponse], error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	rows, err := s.store.ListRBACPermissions(ctx)
	if err != nil {
		s.log.WithError(err).Error("list rbac permissions")
		return nil, mapStoreError(err)
	}
	out := &apiv1.ListRbacPermissionsResponse{}
	for i := range rows {
		out.Permissions = append(out.Permissions, toProtoRbacPermission(&rows[i]))
	}
	return connect.NewResponse(out), nil
}

func (s *Server) ListRbacRoles(ctx context.Context, _ *connect.Request[apiv1.ListRbacRolesRequest]) (*connect.Response[apiv1.ListRbacRolesResponse], error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	rows, err := s.store.ListRBACRoles(ctx)
	if err != nil {
		s.log.WithError(err).Error("list rbac roles")
		return nil, mapStoreError(err)
	}
	out := &apiv1.ListRbacRolesResponse{}
	for i := range rows {
		out.Roles = append(out.Roles, toProtoRbacRole(&rows[i]))
	}
	return connect.NewResponse(out), nil
}

func (s *Server) CreateRbacRole(ctx context.Context, req *connect.Request[apiv1.CreateRbacRoleRequest]) (*connect.Response[apiv1.CreateRbacRoleResponse], error) {
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Msg.Name)
	if name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name required"))
	}
	id, err := s.store.CreateRBACRole(ctx, name, req.Msg.Description, intSlice(req.Msg.PermissionIds))
	if err != nil {
		s.log.WithError(err).Error("create rbac role")
		return nil, mapStoreError(err)
	}
	role, err := s.store.GetRBACRole(ctx, id)
	if err != nil {
		s.log.WithError(err).Error("get created rbac role")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.CreateRbacRoleResponse{Role: toProtoRbacRole(role)}), nil
}

func (s *Server) UpdateRbacRole(ctx context.Context, req *connect.Request[apiv1.UpdateRbacRoleRequest]) (*connect.Response[apiv1.RbacRole], error) {
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if err := s.store.UpdateRBACRole(ctx, int(req.Msg.Id), req.Msg.Name, req.Msg.Description, intSlice(req.Msg.PermissionIds)); err != nil {
		s.log.WithError(err).Error("update rbac role")
		return nil, mapStoreError(err)
	}
	role, err := s.store.GetRBACRole(ctx, int(req.Msg.Id))
	if err != nil {
		s.log.WithError(err).Error("get updated rbac role")
		return nil, mapStoreError(err)
	}
	if role == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("role not found"))
	}
	return connect.NewResponse(toProtoRbacRole(role)), nil
}

func (s *Server) DeleteRbacRole(ctx context.Context, req *connect.Request[apiv1.DeleteRbacRoleRequest]) (*connect.Response[apiv1.DeleteRbacRoleResponse], error) {
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if err := s.store.DeleteRBACRole(ctx, int(req.Msg.Id)); err != nil {
		s.log.WithError(err).Error("delete rbac role")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.DeleteRbacRoleResponse{}), nil
}

func (s *Server) GetUserRbacRoles(ctx context.Context, req *connect.Request[apiv1.GetUserRbacRolesRequest]) (*connect.Response[apiv1.GetUserRbacRolesResponse], error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	names, err := s.store.GetUserRbacRoleNames(ctx, int(req.Msg.UserId))
	if err != nil {
		s.log.WithError(err).Error("get user rbac role names")
		return nil, mapStoreError(err)
	}
	roles, err := s.rbacRolesByNames(ctx, names)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&apiv1.GetUserRbacRolesResponse{Roles: roles}), nil
}

func (s *Server) rbacRolesByNames(ctx context.Context, names []string) ([]*apiv1.RbacRole, error) {
	if len(names) == 0 {
		return nil, nil
	}
	want := make(map[string]struct{}, len(names))
	for _, name := range names {
		want[name] = struct{}{}
	}
	all, err := s.store.ListRBACRoles(ctx)
	if err != nil {
		s.log.WithError(err).Error("list rbac roles")
		return nil, mapStoreError(err)
	}
	out := make([]*apiv1.RbacRole, 0, len(names))
	for i := range all {
		if _, ok := want[all[i].Name]; ok {
			out = append(out, toProtoRbacRole(&all[i]))
		}
	}
	return out, nil
}

func (s *Server) GetUserGroupRbacRoles(ctx context.Context, req *connect.Request[apiv1.GetUserGroupRbacRolesRequest]) (*connect.Response[apiv1.GetUserGroupRbacRolesResponse], error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	roleIDs, err := s.store.GetUserGroupRbacRoleIDs(ctx, int(req.Msg.GroupId))
	if err != nil {
		s.log.WithError(err).Error("get user group rbac role ids")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.GetUserGroupRbacRolesResponse{
		RoleIds: int32Slice(roleIDs),
	}), nil
}

func (s *Server) SetUserGroupRbacRoles(ctx context.Context, req *connect.Request[apiv1.SetUserGroupRbacRolesRequest]) (*connect.Response[apiv1.SetUserGroupRbacRolesResponse], error) {
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if err := s.store.SetUserGroupRbacRoles(ctx, int(req.Msg.GroupId), intSlice(req.Msg.RoleIds)); err != nil {
		s.log.WithError(err).Error("set user group rbac roles")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.SetUserGroupRbacRolesResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Group roles updated"},
	}), nil
}

func (s *Server) GetRbacRoleUsers(ctx context.Context, req *connect.Request[apiv1.GetRbacRoleUsersRequest]) (*connect.Response[apiv1.GetRbacRoleUsersResponse], error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	usernames, err := s.store.ListRbacRoleUsernames(ctx, int(req.Msg.RoleId))
	if err != nil {
		s.log.WithError(err).Error("list rbac role usernames")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.GetRbacRoleUsersResponse{
		Usernames: append([]string(nil), usernames...),
	}), nil
}

func (s *Server) GetRbacRoleGroups(ctx context.Context, req *connect.Request[apiv1.GetRbacRoleGroupsRequest]) (*connect.Response[apiv1.GetRbacRoleGroupsResponse], error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	groupNames, err := s.store.ListRbacRoleGroupNames(ctx, int(req.Msg.RoleId))
	if err != nil {
		s.log.WithError(err).Error("list rbac role group names")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.GetRbacRoleGroupsResponse{
		GroupNames: append([]string(nil), groupNames...),
	}), nil
}

func (s *Server) GetMyPermissionsAudit(ctx context.Context, _ *connect.Request[apiv1.GetMyPermissionsAuditRequest]) (*connect.Response[apiv1.GetMyPermissionsAuditResponse], error) {
	au, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	groupNames, roleNames, isSuperuser, rows, err := s.store.GetMyPermissionsAudit(ctx, au.User.ID)
	if err != nil {
		s.log.WithError(err).Error("get my permissions audit")
		return nil, mapStoreError(err)
	}
	out := &apiv1.GetMyPermissionsAuditResponse{
		GroupNames:  append([]string(nil), groupNames...),
		RoleNames:   append([]string(nil), roleNames...),
		IsSuperuser: isSuperuser,
		Permissions: make([]*apiv1.MyPermissionAuditRow, 0, len(rows)),
	}
	for i := range rows {
		out.Permissions = append(out.Permissions, toProtoMyPermissionAuditRow(&rows[i]))
	}
	return connect.NewResponse(out), nil
}

func (s *Server) ListUserGroups(ctx context.Context, _ *connect.Request[apiv1.ListUserGroupsRequest]) (*connect.Response[apiv1.ListUserGroupsResponse], error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	rows, err := s.store.ListUserGroups(ctx)
	if err != nil {
		s.log.WithError(err).Error("list user groups")
		return nil, mapStoreError(err)
	}
	out := &apiv1.ListUserGroupsResponse{}
	for i := range rows {
		out.Groups = append(out.Groups, toProtoUserGroup(&rows[i]))
	}
	return connect.NewResponse(out), nil
}

func (s *Server) CreateUserGroup(ctx context.Context, req *connect.Request[apiv1.CreateUserGroupRequest]) (*connect.Response[apiv1.CreateUserGroupResponse], error) {
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Msg.Name)
	if name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name required"))
	}
	id, err := s.store.CreateUserGroup(ctx, name)
	if err != nil {
		s.log.WithError(err).Error("create user group")
		return nil, mapStoreError(err)
	}
	group, err := s.store.GetUserGroupByID(ctx, id)
	if err != nil {
		s.log.WithError(err).Error("get created user group")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.CreateUserGroupResponse{Group: toProtoUserGroup(group)}), nil
}

func (s *Server) DeleteUserGroup(ctx context.Context, req *connect.Request[apiv1.DeleteUserGroupRequest]) (*connect.Response[apiv1.DeleteUserGroupResponse], error) {
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if err := s.store.DeleteUserGroup(ctx, int(req.Msg.GroupId)); err != nil {
		s.log.WithError(err).Error("delete user group")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.DeleteUserGroupResponse{}), nil
}

func (s *Server) GetUserGroupMembers(ctx context.Context, req *connect.Request[apiv1.GetUserGroupMembersRequest]) (*connect.Response[apiv1.GetUserGroupMembersResponse], error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	memberIDs, err := s.store.ListUserGroupMemberIDs(ctx, int(req.Msg.GroupId))
	if err != nil {
		s.log.WithError(err).Error("list user group member ids")
		return nil, mapStoreError(err)
	}
	out := &apiv1.GetUserGroupMembersResponse{}
	for _, userID := range memberIDs {
		user, err := s.store.GetUserByID(ctx, userID)
		if err != nil {
			s.log.WithError(err).Error("get user group member")
			return nil, mapStoreError(err)
		}
		if user != nil {
			out.Members = append(out.Members, toProtoUser(user))
		}
	}
	return connect.NewResponse(out), nil
}

func (s *Server) SetUserGroupMembers(ctx context.Context, req *connect.Request[apiv1.SetUserGroupMembersRequest]) (*connect.Response[apiv1.SetUserGroupMembersResponse], error) {
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if err := s.store.SetUserGroupMembers(ctx, int(req.Msg.GroupId), intSlice(req.Msg.UserIds)); err != nil {
		s.log.WithError(err).Error("set user group members")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.SetUserGroupMembersResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Group members updated"},
	}), nil
}

func (s *Server) ImpersonateUser(ctx context.Context, req *connect.Request[apiv1.ImpersonateUserRequest]) (*connect.Response[apiv1.ImpersonateUserResponse], error) {
	au, err := s.requireWrite(ctx)
	if err != nil {
		return nil, err
	}
	targetID := int(req.Msg.UserId)
	target, err := s.store.GetUserByID(ctx, targetID)
	if err != nil {
		s.log.WithError(err).Error("get impersonation target")
		return nil, mapStoreError(err)
	}
	if target == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("user not found"))
	}
	impersonatorID, err := s.impersonatorUserID(ctx, req.Header(), au)
	if err != nil {
		return nil, err
	}
	if impersonatorID == targetID {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("cannot impersonate yourself"))
	}
	oldSID := sessionIDFromHeader(req.Header())
	if oldSID != "" {
		_ = s.store.DeleteSession(ctx, oldSID)
	}
	newSID := uuid.NewString()
	if err := s.store.CreateSession(ctx, newSID, targetID, &impersonatorID); err != nil {
		s.log.WithError(err).Error("create impersonation session")
		return nil, mapStoreError(err)
	}
	res := connect.NewResponse(&apiv1.ImpersonateUserResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Impersonation started"},
	})
	c := auth.NewSessionCookie(newSID)
	res.Header().Add("Set-Cookie", c.String())
	return res, nil
}

func (s *Server) StopImpersonation(ctx context.Context, req *connect.Request[apiv1.StopImpersonationRequest]) (*connect.Response[apiv1.StopImpersonationResponse], error) {
	if _, err := s.requireAuth(ctx); err != nil {
		return nil, err
	}
	sid := sessionIDFromHeader(req.Header())
	if sid == "" {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("not impersonating"))
	}
	sess, err := s.store.GetSessionBySID(ctx, sid)
	if err != nil {
		s.log.WithError(err).Error("get session for stop impersonation")
		return nil, mapStoreError(err)
	}
	if sess == nil || sess.ImpersonatorUserID == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("not impersonating"))
	}
	impersonatorID := *sess.ImpersonatorUserID
	if err := s.store.DeleteSession(ctx, sid); err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.log.WithError(err).Error("delete impersonation session")
		return nil, mapStoreError(err)
	}
	newSID := uuid.NewString()
	if err := s.store.CreateSession(ctx, newSID, impersonatorID, nil); err != nil {
		s.log.WithError(err).Error("restore impersonator session")
		return nil, mapStoreError(err)
	}
	res := connect.NewResponse(&apiv1.StopImpersonationResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Impersonation stopped"},
	})
	c := auth.NewSessionCookie(newSID)
	res.Header().Add("Set-Cookie", c.String())
	return res, nil
}

func (s *Server) ListApiKeys(ctx context.Context, _ *connect.Request[apiv1.ListApiKeysRequest]) (*connect.Response[apiv1.ListApiKeysResponse], error) {
	au, err := s.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.store.ListAPIKeysForUser(ctx, au.User.ID)
	if err != nil {
		s.log.WithError(err).Error("list api keys")
		return nil, mapStoreError(err)
	}
	out := &apiv1.ListApiKeysResponse{}
	for i := range rows {
		out.Keys = append(out.Keys, toProtoApiKey(&rows[i]))
	}
	return connect.NewResponse(out), nil
}

func (s *Server) CreateApiKey(ctx context.Context, req *connect.Request[apiv1.CreateApiKeyRequest]) (*connect.Response[apiv1.CreateApiKeyResponse], error) {
	au, err := s.requireWrite(ctx)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Msg.Name)
	if name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name required"))
	}
	secret, err := password.GenerateAPIKey("sa_")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	id, err := s.store.CreateAPIKey(ctx, au.User.ID, name, secret, req.Msg.ReadOnly)
	if err != nil {
		s.log.WithError(err).Error("create api key")
		return nil, mapStoreError(err)
	}
	rows, err := s.store.ListAPIKeysForUser(ctx, au.User.ID)
	if err != nil {
		s.log.WithError(err).Error("list api keys after create")
		return nil, mapStoreError(err)
	}
	var created *store.APIKeyRow
	for i := range rows {
		if rows[i].ID == id {
			created = &rows[i]
			break
		}
	}
	return connect.NewResponse(&apiv1.CreateApiKeyResponse{
		Key:    toProtoApiKey(created),
		Secret: secret,
	}), nil
}

func (s *Server) RegenerateApiKey(ctx context.Context, req *connect.Request[apiv1.RegenerateApiKeyRequest]) (*connect.Response[apiv1.RegenerateApiKeyResponse], error) {
	au, err := s.requireWrite(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.store.ListAPIKeysForUser(ctx, au.User.ID)
	if err != nil {
		s.log.WithError(err).Error("list api keys before regenerate")
		return nil, mapStoreError(err)
	}
	var old *store.APIKeyRow
	for i := range rows {
		if rows[i].ID == int(req.Msg.Id) {
			old = &rows[i]
			break
		}
	}
	if old == nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("api key not found"))
	}
	secret, err := password.GenerateAPIKey("sa_")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// Revokes the compromised secret.
	if err := s.store.DeleteAPIKey(ctx, old.ID, au.User.ID); err != nil {
		s.log.WithError(err).Error("delete api key during regenerate")
		return nil, mapStoreError(err)
	}
	id, err := s.store.CreateAPIKey(ctx, au.User.ID, old.Name, secret, old.ReadOnly)
	if err != nil {
		s.log.WithError(err).Error("create api key during regenerate")
		return nil, mapStoreError(err)
	}
	rows, err = s.store.ListAPIKeysForUser(ctx, au.User.ID)
	if err != nil {
		s.log.WithError(err).Error("list api keys after regenerate")
		return nil, mapStoreError(err)
	}
	var created *store.APIKeyRow
	for i := range rows {
		if rows[i].ID == id {
			created = &rows[i]
			break
		}
	}
	return connect.NewResponse(&apiv1.RegenerateApiKeyResponse{
		Key:    toProtoApiKey(created),
		Secret: secret,
	}), nil
}

func (s *Server) DeleteApiKey(ctx context.Context, req *connect.Request[apiv1.DeleteApiKeyRequest]) (*connect.Response[apiv1.DeleteApiKeyResponse], error) {
	au, err := s.requireWrite(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.store.DeleteAPIKey(ctx, int(req.Msg.Id), au.User.ID); err != nil {
		s.log.WithError(err).Error("delete api key")
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.DeleteApiKeyResponse{}), nil
}
