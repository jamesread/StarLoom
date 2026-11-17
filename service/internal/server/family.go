package server

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"

	"github.com/jamesread/armature-iam/password"
	apiv1 "github.com/jamesread/starapp/service/gen/starapp/api/v1"
	"github.com/jamesread/starapp/service/internal/avatar"
	"github.com/jamesread/starapp/service/internal/rbac"
	"github.com/jamesread/starapp/service/internal/store"
)

func (s *Server) GetMyFamily(ctx context.Context, _ *connect.Request[apiv1.GetMyFamilyRequest]) (*connect.Response[apiv1.GetMyFamilyResponse], error) {
	fc, err := s.loadFamilyContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.requirePermission(ctx, rbac.PermissionFamilyView); err != nil {
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
	if err = s.requirePermission(ctx, rbac.PermissionFamilyManage); err != nil {
		return nil, err
	}
	return s.createFirstFamily(ctx, au.User.ID, au.User.Username, strings.TrimSpace(req.Msg.Name))
}

func (s *Server) createFirstFamily(ctx context.Context, userID int, username, name string) (*connect.Response[apiv1.CreateFamilyResponse], error) {
	if name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("family name required"))
	}
	if err := s.ensureNoExistingFamily(ctx); err != nil {
		return nil, err
	}
	familyID, memberID, err := s.insertFamilyWithParent(ctx, userID, username, name)
	if err != nil {
		return nil, err
	}
	family, _ := s.store.GetFamilyByID(ctx, familyID)
	member, _ := s.store.GetMemberByID(ctx, memberID)
	return connect.NewResponse(&apiv1.CreateFamilyResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Family created"},
		Family:           toProtoFamily(family),
		CallerMember:     toProtoMember(member),
	}), nil
}

func (s *Server) ensureNoExistingFamily(ctx context.Context) error {
	count, err := s.store.CountFamilies(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	if count > 0 {
		return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("family already exists"))
	}
	return nil
}

func (s *Server) insertFamilyWithParent(ctx context.Context, userID int, username, name string) (familyID, memberID int, err error) {
	familyID, err = s.store.CreateFamily(ctx, name)
	if err != nil {
		return 0, 0, mapStoreError(err)
	}
	if _, err = s.store.CreateStarChart(ctx, familyID, "Star Chart", 0); err != nil {
		return 0, 0, mapStoreError(err)
	}
	memberID, err = s.store.CreateMember(ctx, familyID, username, store.MemberRoleParent, &userID, "")
	if err != nil {
		return 0, 0, mapStoreError(err)
	}
	_ = s.store.EnsureUserInGroup(ctx, userID, rbac.GroupParents)
	return familyID, memberID, nil
}

func (s *Server) ListMembers(ctx context.Context, _ *connect.Request[apiv1.ListMembersRequest]) (*connect.Response[apiv1.ListMembersResponse], error) {
	fc, err := s.requireFamilyPermission(ctx, rbac.PermissionFamilyView)
	if err != nil {
		return nil, err
	}
	rows, err := s.store.ListMembersByFamily(ctx, fc.family.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&apiv1.ListMembersResponse{Members: visibleMembers(fc, rows)}), nil
}

func visibleMembers(fc *familyContext, rows []store.FamilyMemberRow) []*apiv1.FamilyMember {
	out := make([]*apiv1.FamilyMember, 0, len(rows))
	for i := range rows {
		if fc.au.HasPermission(rbac.PermissionStarsViewFamily) || rows[i].ID == fc.member.ID {
			out = append(out, toProtoMember(&rows[i]))
		}
	}
	return out
}

func (s *Server) CreateChildMember(ctx context.Context, req *connect.Request[apiv1.CreateChildMemberRequest]) (*connect.Response[apiv1.CreateChildMemberResponse], error) {
	fc, err := s.requireFamilyWritePermission(ctx, rbac.PermissionMembersManage)
	if err != nil {
		return nil, err
	}
	return s.createChildMember(ctx, fc, req.Msg)
}

func (s *Server) createChildMember(ctx context.Context, fc *familyContext, msg *apiv1.CreateChildMemberRequest) (*connect.Response[apiv1.CreateChildMemberResponse], error) {
	displayName := strings.TrimSpace(msg.DisplayName)
	if displayName == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("display name required"))
	}
	accountID, err := s.optionalChildAccount(ctx, strings.TrimSpace(msg.Username), msg.Password)
	if err != nil {
		return nil, err
	}
	member, err := s.insertChildMember(ctx, fc.family.ID, displayName, accountID)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&apiv1.CreateChildMemberResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Person added"},
		Member:           toProtoMember(member),
	}), nil
}

func (s *Server) optionalChildAccount(ctx context.Context, username, plainPassword string) (*int, error) {
	if plainPassword == "" {
		return nil, nil
	}
	if err := validateOptionalLogin(username, plainPassword); err != nil {
		return nil, err
	}
	return s.hashAndCreateChildUser(ctx, username, plainPassword)
}

func validateOptionalLogin(username, plainPassword string) error {
	if username == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("username required when password is set"))
	}
	if len(plainPassword) < 8 {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("password must be at least 8 characters"))
	}
	return nil
}

func (s *Server) hashAndCreateChildUser(ctx context.Context, username, plainPassword string) (*int, error) {
	hash, err := password.Hash(plainPassword)
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
	return &userID, nil
}

func (s *Server) insertChildMember(ctx context.Context, familyID int, displayName string, accountID *int) (*store.FamilyMemberRow, error) {
	starColor, err := s.nextChildStarColor(ctx, familyID)
	if err != nil {
		return nil, err
	}
	memberID, err := s.store.CreateMember(ctx, familyID, displayName, store.MemberRoleChild, accountID, starColor)
	if err != nil {
		if accountID != nil {
			_ = s.store.DeleteUserAccount(ctx, *accountID)
		}
		return nil, mapStoreError(err)
	}
	member, err := s.store.GetMemberByID(ctx, memberID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return member, nil
}

func (s *Server) nextChildStarColor(ctx context.Context, familyID int) (string, error) {
	members, err := s.store.ListMembersByFamily(ctx, familyID)
	if err != nil {
		return "", connect.NewError(connect.CodeInternal, err)
	}
	childCount := 0
	for _, m := range members {
		if m.Role == store.MemberRoleChild {
			childCount++
		}
	}
	return store.NextChildStarColor(childCount), nil
}

func (s *Server) UpdateMember(ctx context.Context, req *connect.Request[apiv1.UpdateMemberRequest]) (*connect.Response[apiv1.UpdateMemberResponse], error) {
	fc, err := s.requireFamilyWritePermission(ctx, rbac.PermissionMembersManage)
	if err != nil {
		return nil, err
	}
	return s.updateFamilyMember(ctx, fc, req.Msg)
}

func (s *Server) updateFamilyMember(ctx context.Context, fc *familyContext, msg *apiv1.UpdateMemberRequest) (*connect.Response[apiv1.UpdateMemberResponse], error) {
	member, err := s.familyMember(ctx, fc.family.ID, int(msg.MemberId))
	if err != nil {
		return nil, err
	}
	displayName, starColor, err := normalizeMemberUpdate(member, msg.DisplayName, msg.StarColor)
	if err != nil {
		return nil, err
	}
	if err = s.store.UpdateMember(ctx, member.ID, displayName, starColor); err != nil {
		return nil, mapStoreError(err)
	}
	balance, err := s.applyMemberStarAdjustment(ctx, fc, member, int(msg.StarAdjustment), msg.AdjustmentNote)
	if err != nil {
		return nil, err
	}
	member, _ = s.store.GetMemberByID(ctx, member.ID)
	return connect.NewResponse(&apiv1.UpdateMemberResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Member updated"},
		Member:           toProtoMember(member),
		NewBalance:       int32(balance),
	}), nil
}

func normalizeMemberUpdate(member *store.FamilyMemberRow, displayName, starColorRaw string) (string, string, error) {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return "", "", connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("display name required"))
	}
	starColor := member.StarColor
	if c := strings.TrimSpace(starColorRaw); c != "" {
		normalized := store.NormalizeMemberStarColor(c)
		if normalized == "" {
			return "", "", connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("star color must be #RRGGBB"))
		}
		starColor = normalized
	}
	if starColor == "" {
		starColor = store.DefaultMemberStarColor(member.ID)
	}
	return displayName, starColor, nil
}

func (s *Server) AssignMemberLogin(ctx context.Context, req *connect.Request[apiv1.AssignMemberLoginRequest]) (*connect.Response[apiv1.AssignMemberLoginResponse], error) {
	fc, err := s.requireFamilyWritePermission(ctx, rbac.PermissionMembersManage)
	if err != nil {
		return nil, err
	}
	return s.assignMemberLogin(ctx, fc, req.Msg)
}

func (s *Server) assignMemberLogin(ctx context.Context, fc *familyContext, msg *apiv1.AssignMemberLoginRequest) (*connect.Response[apiv1.AssignMemberLoginResponse], error) {
	member, err := s.familyMember(ctx, fc.family.ID, int(msg.MemberId))
	if err != nil {
		return nil, err
	}
	if member.UserAccountID != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("member already has login"))
	}
	userID, err := s.createLoginAccount(ctx, member, strings.TrimSpace(msg.Username), msg.Password)
	if err != nil {
		return nil, err
	}
	if err := s.store.SetMemberUserAccount(ctx, member.ID, userID); err != nil {
		_ = s.store.DeleteUserAccount(ctx, userID)
		return nil, mapStoreError(err)
	}
	member, _ = s.store.GetMemberByID(ctx, member.ID)
	return connect.NewResponse(&apiv1.AssignMemberLoginResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Login created"},
		Member:           toProtoMember(member),
	}), nil
}

func (s *Server) createLoginAccount(ctx context.Context, member *store.FamilyMemberRow, username, plainPassword string) (int, error) {
	if err := validateRequiredLogin(username, plainPassword); err != nil {
		return 0, err
	}
	userID, err := s.createUserWithHash(ctx, username, plainPassword)
	if err != nil {
		return 0, err
	}
	if err := s.assignMemberLoginGroup(ctx, userID, member.Role); err != nil {
		return 0, err
	}
	return userID, nil
}

func validateRequiredLogin(username, plainPassword string) error {
	if username == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("username required"))
	}
	if len(plainPassword) < 8 {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("password must be at least 8 characters"))
	}
	return nil
}

func (s *Server) createUserWithHash(ctx context.Context, username, plainPassword string) (int, error) {
	hash, err := password.Hash(plainPassword)
	if err != nil {
		return 0, connect.NewError(connect.CodeInternal, err)
	}
	userID, err := s.store.CreateUserAccount(ctx, username, hash, store.UserCreatedByAdmin)
	if err != nil {
		return 0, mapStoreError(err)
	}
	return userID, nil
}

func (s *Server) assignMemberLoginGroup(ctx context.Context, userID int, role string) error {
	groupName := rbac.GroupChildren
	if role == store.MemberRoleParent {
		groupName = rbac.GroupParents
	}
	if err := s.store.EnsureUserInGroup(ctx, userID, groupName); err != nil {
		_ = s.store.DeleteUserAccount(ctx, userID)
		return mapStoreError(err)
	}
	return nil
}

func (s *Server) DeleteMember(ctx context.Context, req *connect.Request[apiv1.DeleteMemberRequest]) (*connect.Response[apiv1.DeleteMemberResponse], error) {
	fc, err := s.requireFamilyWritePermission(ctx, rbac.PermissionMembersManage)
	if err != nil {
		return nil, err
	}
	return s.deleteFamilyMember(ctx, fc, int(req.Msg.MemberId))
}

func (s *Server) deleteFamilyMember(ctx context.Context, fc *familyContext, memberID int) (*connect.Response[apiv1.DeleteMemberResponse], error) {
	member, err := s.deletableFamilyMember(ctx, fc, memberID)
	if err != nil {
		return nil, err
	}
	if member.UserAccountID != nil && member.Role == store.MemberRoleChild {
		_ = s.store.DeleteUserAccount(ctx, *member.UserAccountID)
	}
	_ = avatar.DeleteAll(s.cfg.ConfigDir, member.ID)
	if err := s.store.DeleteMember(ctx, member.ID); err != nil {
		return nil, mapStoreError(err)
	}
	return connect.NewResponse(&apiv1.DeleteMemberResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: deleteMemberMessage(member)},
	}), nil
}

func (s *Server) deletableFamilyMember(ctx context.Context, fc *familyContext, memberID int) (*store.FamilyMemberRow, error) {
	member, err := s.familyMemberOrNotFound(ctx, fc.family.ID, memberID)
	if err != nil {
		return nil, err
	}
	if fc.member != nil && fc.member.ID == member.ID {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("cannot remove your own person profile"))
	}
	if err := s.ensureNotLastParent(ctx, fc.family.ID, member); err != nil {
		return nil, err
	}
	return member, nil
}

func (s *Server) familyMemberOrNotFound(ctx context.Context, familyID, memberID int) (*store.FamilyMemberRow, error) {
	member, err := s.store.GetMemberByID(ctx, memberID)
	if err != nil || member == nil || !isFamilyStarMember(member, familyID) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("person not found"))
	}
	return member, nil
}

func (s *Server) ensureNotLastParent(ctx context.Context, familyID int, member *store.FamilyMemberRow) error {
	if member.Role != store.MemberRoleParent {
		return nil
	}
	parentCount, err := s.countParents(ctx, familyID)
	if err != nil {
		return err
	}
	if parentCount <= 1 {
		return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("cannot remove the last parent from the family"))
	}
	return nil
}

func (s *Server) countParents(ctx context.Context, familyID int) (int, error) {
	members, err := s.store.ListMembersByFamily(ctx, familyID)
	if err != nil {
		return 0, connect.NewError(connect.CodeInternal, err)
	}
	parentCount := 0
	for i := range members {
		if members[i].Role == store.MemberRoleParent {
			parentCount++
		}
	}
	return parentCount, nil
}

func deleteMemberMessage(member *store.FamilyMemberRow) string {
	if member.Role == store.MemberRoleParent && member.UserAccountID != nil {
		return "Person removed from the family; their sign-in account was kept"
	}
	return "Person removed from the family"
}

func (s *Server) UploadMemberAvatar(ctx context.Context, req *connect.Request[apiv1.UploadMemberAvatarRequest]) (*connect.Response[apiv1.UploadMemberAvatarResponse], error) {
	fc, err := s.requireFamilyWritePermission(ctx, rbac.PermissionMembersAvatar)
	if err != nil {
		return nil, err
	}
	member, err := s.starFamilyMember(ctx, fc.family.ID, int(req.Msg.MemberId))
	if err != nil {
		return nil, err
	}
	path, err := avatar.Save(s.cfg.ConfigDir, member.ID, req.Msg.Data, req.Msg.ContentType)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return s.setMemberAvatarResponse(ctx, member.ID, path, "Avatar uploaded")
}

func (s *Server) DeleteMemberAvatar(ctx context.Context, req *connect.Request[apiv1.DeleteMemberAvatarRequest]) (*connect.Response[apiv1.DeleteMemberAvatarResponse], error) {
	fc, err := s.requireFamilyWritePermission(ctx, rbac.PermissionMembersAvatar)
	if err != nil {
		return nil, err
	}
	member, err := s.starFamilyMember(ctx, fc.family.ID, int(req.Msg.MemberId))
	if err != nil {
		return nil, err
	}
	_ = avatar.DeleteCurrent(s.cfg.ConfigDir, member.ID)
	memberOut, err := s.persistMemberAvatarPath(ctx, member.ID, "")
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&apiv1.DeleteMemberAvatarResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Avatar removed"},
		Member:           toProtoMember(memberOut),
	}), nil
}

func (s *Server) ListMemberAvatars(ctx context.Context, req *connect.Request[apiv1.ListMemberAvatarsRequest]) (*connect.Response[apiv1.ListMemberAvatarsResponse], error) {
	fc, err := s.requireFamilyPermission(ctx, rbac.PermissionMembersAvatar)
	if err != nil {
		return nil, err
	}
	member, err := s.starFamilyMember(ctx, fc.family.ID, int(req.Msg.MemberId))
	if err != nil {
		return nil, err
	}
	entries, err := avatar.List(s.cfg.ConfigDir, member.ID, member.AvatarPath)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&apiv1.ListMemberAvatarsResponse{Avatars: toProtoAvatarEntries(entries)}), nil
}

func (s *Server) SelectMemberAvatar(ctx context.Context, req *connect.Request[apiv1.SelectMemberAvatarRequest]) (*connect.Response[apiv1.SelectMemberAvatarResponse], error) {
	fc, err := s.requireFamilyWritePermission(ctx, rbac.PermissionMembersAvatar)
	if err != nil {
		return nil, err
	}
	member, err := s.starFamilyMember(ctx, fc.family.ID, int(req.Msg.MemberId))
	if err != nil {
		return nil, err
	}
	path, err := avatar.Select(s.cfg.ConfigDir, member.ID, req.Msg.Filename)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	memberOut, err := s.persistMemberAvatarPath(ctx, member.ID, path)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&apiv1.SelectMemberAvatarResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: "Avatar updated"},
		Member:           toProtoMember(memberOut),
	}), nil
}

func (s *Server) familyMember(ctx context.Context, familyID, memberID int) (*store.FamilyMemberRow, error) {
	member, err := s.store.GetMemberByID(ctx, memberID)
	if err != nil || member == nil || member.FamilyID != familyID {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("member not found"))
	}
	return member, nil
}

func (s *Server) starFamilyMember(ctx context.Context, familyID, memberID int) (*store.FamilyMemberRow, error) {
	member, err := s.familyMember(ctx, familyID, memberID)
	if err != nil {
		return nil, err
	}
	if !isFamilyStarMember(member, familyID) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("member not found"))
	}
	return member, nil
}

func (s *Server) persistMemberAvatarPath(ctx context.Context, memberID int, path string) (*store.FamilyMemberRow, error) {
	if err := s.store.SetMemberAvatarPath(ctx, memberID, path); err != nil {
		return nil, mapStoreError(err)
	}
	member, _ := s.store.GetMemberByID(ctx, memberID)
	return member, nil
}

func (s *Server) setMemberAvatarResponse(ctx context.Context, memberID int, path, message string) (*connect.Response[apiv1.UploadMemberAvatarResponse], error) {
	member, err := s.persistMemberAvatarPath(ctx, memberID, path)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&apiv1.UploadMemberAvatarResponse{
		StandardResponse: &apiv1.StandardResponse{Success: true, Message: message},
		Member:           toProtoMember(member),
	}), nil
}

func toProtoAvatarEntries(entries []avatar.Entry) []*apiv1.MemberAvatarEntry {
	out := make([]*apiv1.MemberAvatarEntry, 0, len(entries))
	for _, entry := range entries {
		out = append(out, &apiv1.MemberAvatarEntry{
			Filename:  entry.Filename,
			IsCurrent: entry.IsCurrent,
		})
	}
	return out
}
