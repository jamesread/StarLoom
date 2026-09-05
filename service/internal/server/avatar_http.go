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
		rest := strings.TrimPrefix(r.URL.Path, "/avatars/")
		rest = strings.TrimSuffix(rest, "/")
		if rest == "" {
			http.NotFound(w, r)
			return
		}

		parts := strings.Split(rest, "/")
		memberID, err := strconv.Atoi(parts[0])
		if err != nil || memberID <= 0 {
			http.NotFound(w, r)
			return
		}

		ctx := r.Context()
		// Image tags cannot send headers.
		if r.Header.Get("Authorization") == "" {
			if token := r.URL.Query().Get("token"); token != "" {
				r.Header.Set("Authorization", "Bearer "+token)
			}
		}
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
		if err != nil || member == nil {
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

		var relativePath string
		switch len(parts) {
		case 1:
			relativePath = member.AvatarPath
		case 2:
			filename := parts[1]
			if !avatar.BelongsToMember(memberID, filename) {
				http.NotFound(w, r)
				return
			}
			relativePath = filename
		default:
			http.NotFound(w, r)
			return
		}
		if relativePath == "" {
			http.NotFound(w, r)
			return
		}

		path := avatar.Path(s.cfg.ConfigDir, relativePath)
		data, err := os.ReadFile(path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		ct := "image/jpeg"
		if strings.HasSuffix(relativePath, ".png") {
			ct = "image/png"
		} else if strings.HasSuffix(relativePath, ".webp") {
			ct = "image/webp"
		}
		w.Header().Set("Content-Type", ct)
		w.Header().Set("Cache-Control", "private, max-age=3600")
		_, _ = w.Write(data)
	})
}
