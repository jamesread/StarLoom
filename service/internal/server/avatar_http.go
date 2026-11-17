package server

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/jamesread/starapp/service/internal/auth"
	"github.com/jamesread/starapp/service/internal/avatar"
	"github.com/jamesread/starapp/service/internal/rbac"
	"github.com/jamesread/starapp/service/internal/store"
)

func (s *Server) AvatarHandler(layer *auth.Layer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.serveAvatar(w, r, layer)
	})
}

func (s *Server) serveAvatar(w http.ResponseWriter, r *http.Request, layer *auth.Layer) {
	memberID, requestedPath, ok := parseAvatarRequest(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	member, ok := s.authorizeAvatarAccess(w, r, layer, memberID)
	if !ok {
		return
	}
	relativePath := requestedPath
	if relativePath == "" {
		relativePath = member.AvatarPath
	}
	if relativePath == "" {
		http.NotFound(w, r)
		return
	}
	writeAvatarBytes(w, s.cfg.ConfigDir, relativePath)
}

func parseAvatarRequest(urlPath string) (memberID int, relativePath string, ok bool) {
	rest := strings.TrimSuffix(strings.TrimPrefix(urlPath, "/avatars/"), "/")
	if rest == "" {
		return 0, "", false
	}
	return parseAvatarParts(strings.Split(rest, "/"))
}

func parseAvatarParts(parts []string) (memberID int, relativePath string, ok bool) {
	memberID, err := strconv.Atoi(parts[0])
	if err != nil || memberID <= 0 {
		return 0, "", false
	}
	switch len(parts) {
	case 1:
		return memberID, "", true
	case 2:
		return avatarFilenamePart(memberID, parts[1])
	default:
		return 0, "", false
	}
}

func avatarFilenamePart(memberID int, filename string) (int, string, bool) {
	if !avatar.BelongsToMember(memberID, filename) {
		return 0, "", false
	}
	return memberID, filename, true
}

func (s *Server) authorizeAvatarAccess(w http.ResponseWriter, r *http.Request, layer *auth.Layer, memberID int) (*store.FamilyMemberRow, bool) {
	au, ok := authenticateAvatarRequest(w, r, layer)
	if !ok {
		return nil, false
	}
	member, err := s.store.GetMemberByID(r.Context(), memberID)
	if err != nil || member == nil {
		http.NotFound(w, r)
		return nil, false
	}
	if !canViewAvatar(r.Context(), s, au, member) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return nil, false
	}
	return member, true
}

func authenticateAvatarRequest(w http.ResponseWriter, r *http.Request, layer *auth.Layer) (*auth.AuthenticatedUser, bool) {
	info, err := layer.Handle(r.Context(), r)
	if err != nil || info == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil, false
	}
	au, ok := info.(*auth.AuthenticatedUser)
	if !ok || au.User == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil, false
	}
	return au, true
}

func canViewAvatar(ctx context.Context, s *Server, au *auth.AuthenticatedUser, member *store.FamilyMemberRow) bool {
	caller, _ := s.store.GetMemberByAccountID(ctx, au.User.ID)
	return avatarVisibleToCaller(au, caller, member)
}

func avatarVisibleToCaller(au *auth.AuthenticatedUser, caller, member *store.FamilyMemberRow) bool {
	if au.HasPermission(rbac.PermissionMembersAvatar) {
		return true
	}
	if caller == nil {
		return false
	}
	if caller.ID == member.ID {
		return true
	}
	return au.HasPermission(rbac.PermissionStarsViewFamily) && caller.FamilyID == member.FamilyID
}

func writeAvatarBytes(w http.ResponseWriter, configDir, relativePath string) {
	path := avatar.Path(configDir, relativePath)
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", avatarContentType(relativePath))
	w.Header().Set("Cache-Control", "private, max-age=3600")
	_, _ = w.Write(data)
}

func avatarContentType(relativePath string) string {
	switch {
	case strings.HasSuffix(relativePath, ".png"):
		return "image/png"
	case strings.HasSuffix(relativePath, ".webp"):
		return "image/webp"
	default:
		return "image/jpeg"
	}
}
