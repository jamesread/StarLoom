package server

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/jamesread/starapp/service/internal/auth"
	"github.com/jamesread/starapp/service/internal/avatar"
	"github.com/jamesread/starapp/service/internal/rbac"
)

func (s *Server) AvatarHandler(layer *auth.Layer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/avatars/")
		idStr = strings.TrimSuffix(idStr, "/")
		memberID, err := strconv.Atoi(idStr)
		if err != nil || memberID <= 0 {
			http.NotFound(w, r)
			return
		}
		ctx := r.Context()
		info, err := layer.Handle(ctx, r)
		if err != nil || info == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		au, ok := info.(*auth.AuthenticatedUser)
		if !ok || au.User == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		member, err := s.store.GetMemberByID(ctx, memberID)
		if err != nil || member == nil || member.AvatarPath == "" {
			http.NotFound(w, r)
			return
		}
		caller, _ := s.store.GetMemberByAccountID(ctx, au.User.ID)
		canView := au.HasPermission(rbac.PermissionStarsViewFamily) &&
			caller != nil && caller.FamilyID == member.FamilyID
		canView = canView || (caller != nil && caller.ID == member.ID)
		if !canView && !au.HasPermission(rbac.PermissionMembersAvatar) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		path := avatar.Path(s.cfg.ConfigDir, member.AvatarPath)
		data, err := os.ReadFile(path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ct := "image/jpeg"
		if strings.HasSuffix(member.AvatarPath, ".png") {
			ct = "image/png"
		} else if strings.HasSuffix(member.AvatarPath, ".webp") {
			ct = "image/webp"
		}
		w.Header().Set("Content-Type", ct)
		w.Header().Set("Cache-Control", "private, max-age=3600")
		_, _ = w.Write(data)
	})
}
