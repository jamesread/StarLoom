package server

import "github.com/jamesread/starapp/service/internal/store"

func isFamilyStarMember(m *store.FamilyMemberRow, familyID int) bool {
	if m == nil || m.FamilyID != familyID {
		return false
	}
	return m.Role == store.MemberRoleChild || m.Role == store.MemberRoleParent
}
