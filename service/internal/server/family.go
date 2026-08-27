package server

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"

	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/avatar"
	"github.com/jamesread/starapp/service/internal/password"
	"github.com/jamesread/starapp/service/internal/rbac"
	"github.com/jamesread/starapp/service/internal/store"
)

func (s *Server) GetMyFamily(ctx context.Context, _ *connect.Request[apiv1.GetMyFamilyRequest]) (*connect.Response[apiv1.GetMyFamilyResponse], error) {
	fc, err := s.loadFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionFamilyView); err != nil {
		return nil, err
	}
	return connect.NewResponse(&apiv1.GetMyFamilyResponse{
		Family:       toProtoFamily(fc.family),
		CallerMember: toProtoMember(fc.member),
	}), nil
}

func (s *Server) CreateFamily(ctx context.Context, req *connect.Request[apiv1.CreateFamilyRequest]) (*connect.Response[apiv1.CreateFamilyResponse], error) {
	au, err := s.requireWrite(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionFamilyManage); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Msg.Name)
	if name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("family name required"))
	}
	count, err := s.store.CountFamilies(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if count > 0 {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("family already exists"))
	}
	familyID, err := s.store.CreateFamily(ctx, name)
	if err != nil {
		return nil, mapStoreError(err)
	}
	memberID, err := s.store.CreateMember(ctx, familyID, au.User.Username, store.MemberRoleParent, &au.User.ID)
	if err != nil {
		return nil, mapStoreError(err)
	}
	_ = s.store.EnsureUserInGroup(ctx, au.User.ID, rbac.GroupParents)
	family, _ := s.store.GetFamilyByID(ctx, familyID)
	member, _ := s.store.GetMemberByID(ctx, memberID)
	return connect.NewResponse(&apiv1.CreateFamilyResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Family created"},
		Family:           toProtoFamily(family),
		CallerMember:     toProtoMember(member),
	}), nil
}

func (s *Server) ListMembers(ctx context.Context, _ *connect.Request[apiv1.ListMembersRequest]) (*connect.Response[apiv1.ListMembersResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if !fc.au.HasPermission(rbac.PermissionFamilyView) {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("forbidden"))
	}
	rows, err := s.store.ListMembersByFamily(ctx, fc.family.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := &apiv1.ListMembersResponse{}
	for i := range rows {
		if fc.au.HasPermission(rbac.PermissionStarsViewFamily) || rows[i].ID == fc.member.ID {
			out.Members = append(out.Members, toProtoMember(&rows[i]))
		}
	}
	return connect.NewResponse(out), nil
}

func (s *Server) CreateChildMember(ctx context.Context, req *connect.Request[apiv1.CreateChildMemberRequest]) (*connect.Response[apiv1.CreateChildMemberResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionMembersManage); err != nil {
		return nil, err
	}
	displayName := strings.TrimSpace(req.Msg.DisplayName)
	username := strings.TrimSpace(req.Msg.Username)
	if displayName == "" || username == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("display name and username required"))
	}
	if len(req.Msg.Password) < 8 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("password must be at least 8 characters"))
	}
	hash, err := password.Hash(req.Msg.Password)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	userID, err := s.store.CreateUserAccount(ctx, username, hash, store.UserCreatedByAdmin)
	if err != nil {
		return nil, mapStoreError(err)
	}
	if err := s.store.EnsureUserInGroup(ctx, userID, rbac.GroupChildren); err != nil {
		return nil, mapStoreError(err)
	}
	memberID, err := s.store.CreateMember(ctx, fc.family.ID, displayName, store.MemberRoleChild, &userID)
	if err != nil {
		_ = s.store.DeleteUserAccount(ctx, userID)
		return nil, mapStoreError(err)
	}
	member, err := s.store.GetMemberByID(ctx, memberID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&apiv1.CreateChildMemberResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Child added"},
		Member:           toProtoMember(member),
	}), nil
}

func (s *Server) UpdateMember(ctx context.Context, req *connect.Request[apiv1.UpdateMemberRequest]) (*connect.Response[apiv1.UpdateMemberResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionMembersManage); err != nil {
		return nil, err
	}
	member, err := s.store.GetMemberByID(ctx, int(req.Msg.MemberId))
	if err != nil || member == nil || member.FamilyID != fc.family.ID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("member not found"))
	}
	displayName := strings.TrimSpace(req.Msg.DisplayName)
	if displayName == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("display name required"))
	}
	if err := s.store.UpdateMemberDisplayName(ctx, member.ID, displayName); err != nil {
		return nil, mapStoreError(err)
	}
	member, _ = s.store.GetMemberByID(ctx, member.ID)
	return connect.NewResponse(&apiv1.UpdateMemberResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Member updated"},
		Member:           toProtoMember(member),
	}), nil
}

func (s *Server) DeleteMember(ctx context.Context, req *connect.Request[apiv1.DeleteMemberRequest]) (*connect.Response[apiv1.DeleteMemberResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionMembersManage); err != nil {
		return nil, err
	}
	member, err := s.store.GetMemberByID(ctx, int(req.Msg.MemberId))
	if err != nil || member == nil || member.FamilyID != fc.family.ID || member.Role != store.MemberRoleChild {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("member not found"))
	}
	if member.UserAccountID != nil {
		_ = s.store.DeleteUserAccount(ctx, *member.UserAccountID)
	}
	_ = avatar.Delete(s.cfg.ConfigDir, member.ID)
	if err := s.store.DeleteMember(ctx, member.ID); err != nil {
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.DeleteMemberResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Member deleted"},
	}), nil
}

func (s *Server) UploadMemberAvatar(ctx context.Context, req *connect.Request[apiv1.UploadMemberAvatarRequest]) (*connect.Response[apiv1.UploadMemberAvatarResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionMembersAvatar); err != nil {
		return nil, err
	}
	member, err := s.store.GetMemberByID(ctx, int(req.Msg.MemberId))
	if err != nil || member == nil || member.FamilyID != fc.family.ID || member.Role != store.MemberRoleChild {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("member not found"))
	}
	path, err := avatar.Save(s.cfg.ConfigDir, member.ID, req.Msg.Data, req.Msg.ContentType)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if err := s.store.SetMemberAvatarPath(ctx, member.ID, path); err != nil {
		return nil, mapStoreError(err)
	}
	member, _ = s.store.GetMemberByID(ctx, member.ID)
	return connect.NewResponse(&apiv1.UploadMemberAvatarResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Avatar uploaded"},
		Member:           toProtoMember(member),
	}), nil
}

func (s *Server) DeleteMemberAvatar(ctx context.Context, req *connect.Request[apiv1.DeleteMemberAvatarRequest]) (*connect.Response[apiv1.DeleteMemberAvatarResponse], error) {
	fc, err := s.requireFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireWrite(ctx); err != nil {
		return nil, err
	}
	if _, err := s.requirePermission(ctx, rbac.PermissionMembersAvatar); err != nil {
		return nil, err
	}
	member, err := s.store.GetMemberByID(ctx, int(req.Msg.MemberId))
	if err != nil || member == nil || member.FamilyID != fc.family.ID || member.Role != store.MemberRoleChild {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("member not found"))
	}
	_ = avatar.Delete(s.cfg.ConfigDir, member.ID)
	if err := s.store.SetMemberAvatarPath(ctx, member.ID, ""); err != nil {
		return nil, mapStoreError(err)
	}
	member, _ = s.store.GetMemberByID(ctx, member.ID)
	return connect.NewResponse(&apiv1.DeleteMemberAvatarResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Avatar removed"},
		Member:           toProtoMember(member),
	}), nil
}
