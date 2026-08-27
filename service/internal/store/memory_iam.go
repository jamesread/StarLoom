package store

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jamesread/starapp/service/internal/rbac"
)

type memoryIAM struct {
	mu sync.Mutex

	seeded bool

	nextUserID    int
	users         map[int]UserAccountRow
	usernameIndex map[string]int

	nextSessionID int
	sessions      map[string]SessionRow

	nextAPIKeyID int
	apiKeys      map[int]APIKeyRow
	apiKeyIndex  map[string]int

	permissions map[int]RBACPermissionRow
	permByName  map[string]int

	roles      map[int]memoryRole
	nextRoleID int

	rolePerms map[int]map[int]struct{}

	groups      map[int]memoryGroup
	groupByName map[string]int
	nextGroupID int

	groupRoles  map[int]map[int]struct{}
	memberships map[int]map[int]struct{}

	userPrefs map[int]UserPreferencesRow
}

type memoryRole struct {
	ID          int
	Name        string
	Description string
	CreatedAt   string
	UpdatedAt   string
}

type memoryGroup struct {
	ID        int
	Name      string
	CreatedAt string
	UpdatedAt string
}

func iamNow() string {
	return time.Now().UTC().Format("2006-01-02 15:04:05")
}

func usernameKey(username string) string {
	return strings.ToLower(username)
}

func (m *Memory) iamState() *memoryIAM {
	if m.iam == nil {
		m.iam = &memoryIAM{
			users:         map[int]UserAccountRow{},
			usernameIndex: map[string]int{},
			sessions:      map[string]SessionRow{},
			apiKeys:       map[int]APIKeyRow{},
			apiKeyIndex:   map[string]int{},
			permissions:   map[int]RBACPermissionRow{},
			permByName:    map[string]int{},
			roles:         map[int]memoryRole{},
			rolePerms:     map[int]map[int]struct{}{},
			groups:        map[int]memoryGroup{},
			groupByName:   map[string]int{},
			groupRoles:    map[int]map[int]struct{}{},
			memberships:   map[int]map[int]struct{}{},
		}
		m.iam.seedRBACCatalog()
	}
	return m.iam
}

func (st *memoryIAM) seedRBACCatalog() {
	if st.seeded {
		return
	}
	st.seeded = true
	now := iamNow()

	permDefs := []RBACPermissionRow{
		{Name: rbac.PermissionAppAccess, Description: "Use the application (non-admin APIs)"},
		{Name: rbac.PermissionUsersView, Description: "List and view user accounts"},
		{Name: rbac.PermissionUsersCreate, Description: "Create user accounts"},
		{Name: rbac.PermissionUsersDelete, Description: "Delete user accounts"},
		{Name: rbac.PermissionUsersResetPassword, Description: "Reset passwords for other users"},
		{Name: rbac.PermissionUserGroupsView, Description: "List user groups and view membership"},
		{Name: rbac.PermissionUserGroupsManage, Description: "Create and delete user groups; manage membership"},
		{Name: rbac.PermissionRbacView, Description: "View roles and permissions"},
		{Name: rbac.PermissionRbacManage, Description: "Create, update, and delete roles; assign roles to groups"},
		{Name: rbac.PermissionSystemSettings, Description: "View and modify system settings"},
		{Name: rbac.PermissionSystemLogs, Description: "View logs, job status, and run maintenance"},
		{Name: rbac.PermissionSystemImpersonate, Description: "Impersonate other users"},
		{Name: rbac.PermissionFamilyView, Description: "View family name and member list"},
		{Name: rbac.PermissionFamilyManage, Description: "Create and update family"},
		{Name: rbac.PermissionMembersManage, Description: "Add, edit, and remove child members"},
		{Name: rbac.PermissionMembersAvatar, Description: "Upload and remove child avatars"},
		{Name: rbac.PermissionStarsViewFamily, Description: "View all children balances and history"},
		{Name: rbac.PermissionStarsViewOwn, Description: "View own balance and history"},
		{Name: rbac.PermissionStarsAward, Description: "Award stars to children"},
		{Name: rbac.PermissionStarsRevoke, Description: "Revoke or correct star entries"},
		{Name: rbac.PermissionRewardsManage, Description: "Manage reward catalog"},
		{Name: rbac.PermissionRewardsView, Description: "Browse active rewards"},
		{Name: rbac.PermissionRedemptionsApprove, Description: "Approve or reject redemptions"},
		{Name: rbac.PermissionRedemptionsRequest, Description: "Request reward redemptions"},
	}
	for i, p := range permDefs {
		id := i + 1
		p.ID = id
		st.permissions[id] = p
		st.permByName[p.Name] = id
	}

	st.roles[1] = memoryRole{ID: 1, Name: rbac.RoleSuperuser, Description: "All permissions (system role)", CreatedAt: now, UpdatedAt: now}
	st.roles[2] = memoryRole{ID: 2, Name: rbac.RoleMember, Description: "Standard application access", CreatedAt: now, UpdatedAt: now}
	st.roles[3] = memoryRole{ID: 3, Name: rbac.RoleParent, Description: "Family administrator", CreatedAt: now, UpdatedAt: now}
	st.roles[4] = memoryRole{ID: 4, Name: rbac.RoleChild, Description: "Child account with limited access", CreatedAt: now, UpdatedAt: now}
	st.nextRoleID = 5

	st.rolePerms[1] = map[int]struct{}{}
	for id := range st.permissions {
		st.rolePerms[1][id] = struct{}{}
	}
	st.rolePerms[2] = map[int]struct{}{1: {}}
	st.rolePerms[3] = map[int]struct{}{}
	parentPerms := []string{
		rbac.PermissionAppAccess, rbac.PermissionFamilyView, rbac.PermissionFamilyManage,
		rbac.PermissionMembersManage, rbac.PermissionMembersAvatar, rbac.PermissionStarsViewFamily,
		rbac.PermissionStarsViewOwn, rbac.PermissionStarsAward, rbac.PermissionStarsRevoke,
		rbac.PermissionRewardsManage, rbac.PermissionRewardsView, rbac.PermissionRedemptionsApprove,
		rbac.PermissionRedemptionsRequest,
	}
	for _, name := range parentPerms {
		if pid, ok := st.permByName[name]; ok {
			st.rolePerms[3][pid] = struct{}{}
		}
	}
	st.rolePerms[4] = map[int]struct{}{}
	childPerms := []string{
		rbac.PermissionAppAccess, rbac.PermissionStarsViewOwn, rbac.PermissionRewardsView,
		rbac.PermissionRedemptionsRequest,
	}
	for _, name := range childPerms {
		if pid, ok := st.permByName[name]; ok {
			st.rolePerms[4][pid] = struct{}{}
		}
	}

	st.groups[1] = memoryGroup{ID: 1, Name: rbac.GroupEveryone, CreatedAt: now, UpdatedAt: now}
	st.groups[2] = memoryGroup{ID: 2, Name: rbac.GroupAdministrators, CreatedAt: now, UpdatedAt: now}
	st.groups[3] = memoryGroup{ID: 3, Name: rbac.GroupParents, CreatedAt: now, UpdatedAt: now}
	st.groups[4] = memoryGroup{ID: 4, Name: rbac.GroupChildren, CreatedAt: now, UpdatedAt: now}
	st.groupByName[rbac.GroupEveryone] = 1
	st.groupByName[rbac.GroupAdministrators] = 2
	st.groupByName[rbac.GroupParents] = 3
	st.groupByName[rbac.GroupChildren] = 4
	st.nextGroupID = 5

	st.groupRoles[1] = map[int]struct{}{2: {}}
	st.groupRoles[2] = map[int]struct{}{1: {}}
	st.groupRoles[3] = map[int]struct{}{3: {}}
	st.groupRoles[4] = map[int]struct{}{4: {}}
	st.memberships[1] = map[int]struct{}{}
	st.memberships[2] = map[int]struct{}{}
	st.memberships[3] = map[int]struct{}{}
	st.memberships[4] = map[int]struct{}{}
}

func (st *memoryIAM) countUsersLocked() int {
	return len(st.users)
}

func (st *memoryIAM) roleIDByNameLocked(name string) (int, bool) {
	for id, r := range st.roles {
		if r.Name == name {
			return id, true
		}
	}
	return 0, false
}

func (st *memoryIAM) countUsersWithSuperuserViaGroupsLocked() int {
	superID, ok := st.roleIDByNameLocked(rbac.RoleSuperuser)
	if !ok {
		return 0
	}
	seen := map[int]struct{}{}
	for groupID, roleSet := range st.groupRoles {
		if _, hasSuper := roleSet[superID]; !hasSuper {
			continue
		}
		for userID := range st.memberships[groupID] {
			seen[userID] = struct{}{}
		}
	}
	return len(seen)
}

func (st *memoryIAM) ensureSuperuserCoverageLocked() error {
	if st.countUsersWithSuperuserViaGroupsLocked() == 0 {
		return errNoSuperuser
	}
	return nil
}

func (st *memoryIAM) deleteUserCascadeLocked(id int) {
	delete(st.users, id)
	for k, uid := range st.usernameIndex {
		if uid == id {
			delete(st.usernameIndex, k)
		}
	}
	for sid, sess := range st.sessions {
		if sess.UserAccountID == id {
			delete(st.sessions, sid)
		}
	}
	for keyID, key := range st.apiKeys {
		if key.UserAccountID == id {
			delete(st.apiKeys, keyID)
			delete(st.apiKeyIndex, key.KeyValue)
		}
	}
	for groupID := range st.memberships {
		delete(st.memberships[groupID], id)
	}
}

func (st *memoryIAM) loadRBACRoleLocked(id int) (*RBACRoleRow, error) {
	role, ok := st.roles[id]
	if !ok {
		return nil, nil
	}
	out := &RBACRoleRow{
		ID:            role.ID,
		Name:          role.Name,
		Description:   role.Description,
		PermissionIDs: st.listRolePermissionIDsLocked(id),
	}
	for _, roleSet := range st.groupRoles {
		if _, ok := roleSet[id]; ok {
			out.GroupCount++
		}
	}
	seenUsers := map[int]struct{}{}
	for groupID, roleSet := range st.groupRoles {
		if _, ok := roleSet[id]; !ok {
			continue
		}
		for userID := range st.memberships[groupID] {
			seenUsers[userID] = struct{}{}
		}
	}
	out.UserCount = len(seenUsers)
	return out, nil
}

func (st *memoryIAM) listRolePermissionIDsLocked(roleID int) []int {
	set := st.rolePerms[roleID]
	out := make([]int, 0, len(set))
	for pid := range set {
		out = append(out, pid)
	}
	sort.Ints(out)
	return out
}

func (st *memoryIAM) setRolePermissionsLocked(roleID int, permissionIDs []int) {
	next := map[int]struct{}{}
	for _, pid := range permissionIDs {
		next[pid] = struct{}{}
	}
	st.rolePerms[roleID] = next
}

func (st *memoryIAM) userGroupRowLocked(id int) (*UserGroupRow, error) {
	g, ok := st.groups[id]
	if !ok {
		return nil, nil
	}
	memberCount := 0
	if members, ok := st.memberships[id]; ok {
		memberCount = len(members)
	}
	return &UserGroupRow{
		ID:          g.ID,
		Name:        g.Name,
		MemberCount: memberCount,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}, nil
}

func (st *memoryIAM) userHasAnyMembershipLocked(userID int) bool {
	for _, members := range st.memberships {
		if _, ok := members[userID]; ok {
			return true
		}
	}
	return false
}

func (st *memoryIAM) minUserIDLocked() (int, bool) {
	minID := 0
	for id := range st.users {
		if minID == 0 || id < minID {
			minID = id
		}
	}
	return minID, minID != 0
}

func (m *Memory) CountUserAccounts(_ context.Context) (int, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	return st.countUsersLocked(), nil
}

func (m *Memory) GetUserByUsername(_ context.Context, username string) (*UserAccountRow, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	id, ok := st.usernameIndex[usernameKey(username)]
	if !ok {
		return nil, nil
	}
	u, ok := st.users[id]
	if !ok {
		return nil, nil
	}
	copy := u
	return &copy, nil
}

func (m *Memory) GetUserByID(_ context.Context, id int) (*UserAccountRow, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	u, ok := st.users[id]
	if !ok {
		return nil, nil
	}
	copy := u
	return &copy, nil
}

func (m *Memory) ListUserAccounts(_ context.Context) ([]UserAccountRow, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	out := make([]UserAccountRow, 0, len(st.users))
	for _, u := range st.users {
		out = append(out, u)
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].Username) < strings.ToLower(out[j].Username)
	})
	return out, nil
}

func (m *Memory) CreateUserAccount(_ context.Context, username, passwordHash, createdBy string) (int, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	if _, ok := st.usernameIndex[usernameKey(username)]; ok {
		return 0, fmt.Errorf("create user account: username already exists")
	}
	st.nextUserID++
	now := iamNow()
	id := st.nextUserID
	st.users[id] = UserAccountRow{
		ID: id, Username: username, PasswordHash: passwordHash,
		CreatedBy: createdBy, CreatedAt: now, UpdatedAt: now,
	}
	st.usernameIndex[usernameKey(username)] = id
	return id, nil
}

func (m *Memory) DeleteUserAccount(_ context.Context, id int) error {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	if _, ok := st.users[id]; !ok {
		return sql.ErrNoRows
	}
	st.deleteUserCascadeLocked(id)
	return nil
}

func (m *Memory) UpdateUserPassword(_ context.Context, id int, passwordHash string) error {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	u, ok := st.users[id]
	if !ok {
		return sql.ErrNoRows
	}
	u.PasswordHash = passwordHash
	u.UpdatedAt = iamNow()
	st.users[id] = u
	return nil
}

func (m *Memory) CreateSession(_ context.Context, sid string, userID int, impersonatorID *int) error {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	st.nextSessionID++
	now := iamNow()
	st.sessions[sid] = SessionRow{
		ID: st.nextSessionID, SID: sid, UserAccountID: userID,
		ImpersonatorUserID: impersonatorID, CreatedAt: now, UpdatedAt: now,
	}
	return nil
}

func (m *Memory) GetSessionBySID(_ context.Context, sid string) (*SessionRow, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	s, ok := st.sessions[sid]
	if !ok {
		return nil, nil
	}
	copy := s
	return &copy, nil
}

func (m *Memory) DeleteSession(_ context.Context, sid string) error {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	if _, ok := st.sessions[sid]; !ok {
		return sql.ErrNoRows
	}
	delete(st.sessions, sid)
	return nil
}

func (m *Memory) DeleteSessionsForUser(_ context.Context, userID int) error {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	for sid, sess := range st.sessions {
		if sess.UserAccountID == userID {
			delete(st.sessions, sid)
		}
	}
	return nil
}

func (m *Memory) ListAPIKeysForUser(_ context.Context, userID int) ([]APIKeyRow, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	out := make([]APIKeyRow, 0)
	for _, k := range st.apiKeys {
		if k.UserAccountID == userID {
			out = append(out, k)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func (m *Memory) CreateAPIKey(_ context.Context, userID int, name, keyValue string, readOnly bool) (int, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	st.nextAPIKeyID++
	now := iamNow()
	id := st.nextAPIKeyID
	st.apiKeys[id] = APIKeyRow{
		ID: id, UserAccountID: userID, Name: name, KeyValue: keyValue,
		ReadOnly: readOnly, CreatedAt: now, UpdatedAt: now,
	}
	st.apiKeyIndex[keyValue] = id
	return id, nil
}

func (m *Memory) DeleteAPIKey(_ context.Context, id, userID int) error {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	k, ok := st.apiKeys[id]
	if !ok || k.UserAccountID != userID {
		return sql.ErrNoRows
	}
	delete(st.apiKeys, id)
	delete(st.apiKeyIndex, k.KeyValue)
	return nil
}

func (m *Memory) GetUserByAPIKey(_ context.Context, keyValue string) (*UserAccountRow, bool, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	keyID, ok := st.apiKeyIndex[keyValue]
	if !ok {
		return nil, false, nil
	}
	k, ok := st.apiKeys[keyID]
	if !ok {
		return nil, false, nil
	}
	u, ok := st.users[k.UserAccountID]
	if !ok {
		return nil, false, nil
	}
	copy := u
	return &copy, k.ReadOnly, nil
}

func (m *Memory) TouchAPIKeyUsed(_ context.Context, keyValue string) error {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	keyID, ok := st.apiKeyIndex[keyValue]
	if !ok {
		return sql.ErrNoRows
	}
	k, ok := st.apiKeys[keyID]
	if !ok {
		return sql.ErrNoRows
	}
	now := iamNow()
	k.LastUsedAt = now
	k.UpdatedAt = now
	st.apiKeys[keyID] = k
	return nil
}

func (m *Memory) LoadEffectiveRBAC(_ context.Context, userID int) (*rbac.EffectiveRBAC, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()

	superID, _ := st.roleIDByNameLocked(rbac.RoleSuperuser)
	isSuperuser := false
	for groupID, roleSet := range st.groupRoles {
		if _, ok := roleSet[superID]; !ok {
			continue
		}
		if members, ok := st.memberships[groupID]; ok {
			if _, ok := members[userID]; ok {
				isSuperuser = true
				break
			}
		}
	}

	out := &rbac.EffectiveRBAC{
		IsSuperuser: isSuperuser,
		Permissions: map[string]bool{},
	}

	if isSuperuser {
		names := make([]string, 0, len(st.permissions))
		for _, p := range st.permissions {
			names = append(names, p.Name)
		}
		sort.Strings(names)
		for _, name := range names {
			out.Permissions[name] = true
		}
	} else {
		permNames := map[string]struct{}{}
		for groupID, roleSet := range st.groupRoles {
			members, ok := st.memberships[groupID]
			if !ok {
				continue
			}
			if _, ok := members[userID]; !ok {
				continue
			}
			for roleID := range roleSet {
				for pid := range st.rolePerms[roleID] {
					if p, ok := st.permissions[pid]; ok {
						permNames[p.Name] = struct{}{}
					}
				}
			}
		}
		names := make([]string, 0, len(permNames))
		for name := range permNames {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			out.Permissions[name] = true
		}
	}

	roleNamesSet := map[string]struct{}{}
	for groupID, roleSet := range st.groupRoles {
		members, ok := st.memberships[groupID]
		if !ok {
			continue
		}
		if _, ok := members[userID]; !ok {
			continue
		}
		for roleID := range roleSet {
			if r, ok := st.roles[roleID]; ok {
				roleNamesSet[r.Name] = struct{}{}
			}
		}
	}
	for name := range roleNamesSet {
		out.RoleNames = append(out.RoleNames, name)
	}
	sort.Strings(out.RoleNames)
	return out, nil
}

func (m *Memory) EnsureRBACBootstrap(_ context.Context) error {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()

	if st.countUsersLocked() == 0 {
		return nil
	}

	everyoneID := st.groupByName[rbac.GroupEveryone]
	adminID := st.groupByName[rbac.GroupAdministrators]
	memberRoleID, _ := st.roleIDByNameLocked(rbac.RoleMember)
	superRoleID, _ := st.roleIDByNameLocked(rbac.RoleSuperuser)

	if st.groupRoles[everyoneID] == nil {
		st.groupRoles[everyoneID] = map[int]struct{}{}
	}
	st.groupRoles[everyoneID][memberRoleID] = struct{}{}

	if st.groupRoles[adminID] == nil {
		st.groupRoles[adminID] = map[int]struct{}{}
	}
	st.groupRoles[adminID][superRoleID] = struct{}{}

	if st.countUsersWithSuperuserViaGroupsLocked() == 0 {
		if minID, ok := st.minUserIDLocked(); ok {
			if st.memberships[adminID] == nil {
				st.memberships[adminID] = map[int]struct{}{}
			}
			st.memberships[adminID][minID] = struct{}{}
		}
	}

	if st.memberships[everyoneID] == nil {
		st.memberships[everyoneID] = map[int]struct{}{}
	}
	for userID := range st.users {
		if !st.userHasAnyMembershipLocked(userID) {
			st.memberships[everyoneID][userID] = struct{}{}
		}
	}

	return nil
}

func (m *Memory) EnsureUserInEveryoneGroup(_ context.Context, userID int) error {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	everyoneID := st.groupByName[rbac.GroupEveryone]
	if st.memberships[everyoneID] == nil {
		st.memberships[everyoneID] = map[int]struct{}{}
	}
	st.memberships[everyoneID][userID] = struct{}{}
	return nil
}

func (m *Memory) CountUsersWithSuperuserViaGroups(_ context.Context) (int, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	return st.countUsersWithSuperuserViaGroupsLocked(), nil
}

func (m *Memory) ListRBACPermissions(_ context.Context) ([]RBACPermissionRow, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	out := make([]RBACPermissionRow, 0, len(st.permissions))
	for _, p := range st.permissions {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (m *Memory) ListRBACRoles(_ context.Context) ([]RBACRoleRow, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	ids := make([]int, 0, len(st.roles))
	for id := range st.roles {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return st.roles[ids[i]].Name < st.roles[ids[j]].Name
	})
	out := make([]RBACRoleRow, 0, len(ids))
	for _, id := range ids {
		role, err := st.loadRBACRoleLocked(id)
		if err != nil {
			return nil, err
		}
		if role != nil {
			out = append(out, *role)
		}
	}
	return out, nil
}

func (m *Memory) GetRBACRole(_ context.Context, id int) (*RBACRoleRow, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	return st.loadRBACRoleLocked(id)
}

func (m *Memory) CreateRBACRole(_ context.Context, name, description string, permissionIDs []int) (int, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	now := iamNow()
	id := st.nextRoleID
	st.nextRoleID++
	st.roles[id] = memoryRole{ID: id, Name: name, Description: description, CreatedAt: now, UpdatedAt: now}
	st.setRolePermissionsLocked(id, permissionIDs)
	return id, nil
}

func (m *Memory) UpdateRBACRole(_ context.Context, id int, name, description string, permissionIDs []int) error {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	cur, ok := st.roles[id]
	if !ok {
		return sql.ErrNoRows
	}
	curName := cur.Name
	if isSystemRBACRole(curName) && name != curName {
		return fmt.Errorf("cannot rename system role %q", curName)
	}
	cur.Name = name
	cur.Description = description
	cur.UpdatedAt = iamNow()
	st.roles[id] = cur
	if curName != rbac.RoleSuperuser {
		st.setRolePermissionsLocked(id, permissionIDs)
	}
	return nil
}

func (m *Memory) DeleteRBACRole(_ context.Context, id int) error {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	cur, ok := st.roles[id]
	if !ok {
		return sql.ErrNoRows
	}
	if isSystemRBACRole(cur.Name) {
		return fmt.Errorf("cannot delete system role %q", cur.Name)
	}
	delete(st.roles, id)
	delete(st.rolePerms, id)
	for gid, roleSet := range st.groupRoles {
		delete(roleSet, id)
		if len(roleSet) == 0 {
			delete(st.groupRoles, gid)
		}
	}
	return nil
}

func (m *Memory) SetRBACRolePermissions(_ context.Context, roleID int, permissionIDs []int) error {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	cur, ok := st.roles[roleID]
	if !ok {
		return sql.ErrNoRows
	}
	if cur.Name == rbac.RoleSuperuser {
		return fmt.Errorf("cannot set permissions for system role %q", cur.Name)
	}
	st.setRolePermissionsLocked(roleID, permissionIDs)
	return nil
}

func (m *Memory) ListRolePermissionIDs(_ context.Context, roleID int) ([]int, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	if _, ok := st.roles[roleID]; !ok {
		return []int{}, nil
	}
	return st.listRolePermissionIDsLocked(roleID), nil
}

func (m *Memory) ListPermissionRoleNames(_ context.Context, permissionID int) ([]string, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	names := make([]string, 0)
	for roleID, role := range st.roles {
		if perms, ok := st.rolePerms[roleID]; ok {
			if _, ok := perms[permissionID]; ok {
				names = append(names, role.Name)
			}
		}
	}
	sort.Strings(names)
	return names, nil
}

func (m *Memory) GetUserRbacRoleNames(_ context.Context, userID int) ([]string, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	namesSet := map[string]struct{}{}
	for groupID, roleSet := range st.groupRoles {
		members, ok := st.memberships[groupID]
		if !ok {
			continue
		}
		if _, ok := members[userID]; !ok {
			continue
		}
		for roleID := range roleSet {
			if r, ok := st.roles[roleID]; ok {
				namesSet[r.Name] = struct{}{}
			}
		}
	}
	names := make([]string, 0, len(namesSet))
	for name := range namesSet {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

func (m *Memory) GetUserGroupRbacRoleIDs(_ context.Context, groupID int) ([]int, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	roleSet := st.groupRoles[groupID]
	out := make([]int, 0, len(roleSet))
	for roleID := range roleSet {
		out = append(out, roleID)
	}
	sort.Ints(out)
	return out, nil
}

func (m *Memory) SetUserGroupRbacRoles(_ context.Context, groupID int, roleIDs []int) error {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	if _, ok := st.groups[groupID]; !ok {
		return sql.ErrNoRows
	}
	next := map[int]struct{}{}
	for _, roleID := range roleIDs {
		next[roleID] = struct{}{}
	}
	st.groupRoles[groupID] = next
	return st.ensureSuperuserCoverageLocked()
}

func (m *Memory) ListRbacRoleGroupNames(_ context.Context, roleID int) ([]string, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	names := make([]string, 0)
	for groupID, roleSet := range st.groupRoles {
		if _, ok := roleSet[roleID]; !ok {
			continue
		}
		if g, ok := st.groups[groupID]; ok {
			names = append(names, g.Name)
		}
	}
	sort.Strings(names)
	return names, nil
}

func (m *Memory) ListRbacRoleUsernames(_ context.Context, roleID int) ([]string, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	seen := map[int]struct{}{}
	for groupID, roleSet := range st.groupRoles {
		if _, ok := roleSet[roleID]; !ok {
			continue
		}
		for userID := range st.memberships[groupID] {
			seen[userID] = struct{}{}
		}
	}
	names := make([]string, 0, len(seen))
	for userID := range seen {
		if u, ok := st.users[userID]; ok {
			names = append(names, u.Username)
		}
	}
	sort.Slice(names, func(i, j int) bool {
		return strings.ToLower(names[i]) < strings.ToLower(names[j])
	})
	return names, nil
}

func (m *Memory) GetMyPermissionsAudit(_ context.Context, userID int) ([]string, []string, bool, []MyPermissionAuditRow, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()

	groupNames := make([]string, 0)
	for groupID, members := range st.memberships {
		if _, ok := members[userID]; !ok {
			continue
		}
		if g, ok := st.groups[groupID]; ok {
			groupNames = append(groupNames, g.Name)
		}
	}
	sort.Strings(groupNames)

	roleNamesSet := map[string]struct{}{}
	for groupID, roleSet := range st.groupRoles {
		members, ok := st.memberships[groupID]
		if !ok {
			continue
		}
		if _, ok := members[userID]; !ok {
			continue
		}
		for roleID := range roleSet {
			if r, ok := st.roles[roleID]; ok {
				roleNamesSet[r.Name] = struct{}{}
			}
		}
	}
	roleNames := make([]string, 0, len(roleNamesSet))
	for name := range roleNamesSet {
		roleNames = append(roleNames, name)
	}
	sort.Strings(roleNames)

	superID, _ := st.roleIDByNameLocked(rbac.RoleSuperuser)
	isSuperuser := false
	for groupID, roleSet := range st.groupRoles {
		if _, ok := roleSet[superID]; !ok {
			continue
		}
		if members, ok := st.memberships[groupID]; ok {
			if _, ok := members[userID]; ok {
				isSuperuser = true
				break
			}
		}
	}

	effectivePerms := map[string]bool{}
	if isSuperuser {
		for _, p := range st.permissions {
			effectivePerms[p.Name] = true
		}
	} else {
		for groupID, roleSet := range st.groupRoles {
			members, ok := st.memberships[groupID]
			if !ok {
				continue
			}
			if _, ok := members[userID]; !ok {
				continue
			}
			for roleID := range roleSet {
				for pid := range st.rolePerms[roleID] {
					if p, ok := st.permissions[pid]; ok {
						effectivePerms[p.Name] = true
					}
				}
			}
		}
	}

	permCatalog := make([]RBACPermissionRow, 0, len(st.permissions))
	for _, p := range st.permissions {
		permCatalog = append(permCatalog, p)
	}
	sort.Slice(permCatalog, func(i, j int) bool { return permCatalog[i].Name < permCatalog[j].Name })

	grantingByPerm := map[string][]string{}
	if !isSuperuser {
		for groupID, roleSet := range st.groupRoles {
			members, ok := st.memberships[groupID]
			if !ok {
				continue
			}
			if _, ok := members[userID]; !ok {
				continue
			}
			g, ok := st.groups[groupID]
			if !ok {
				continue
			}
			for roleID := range roleSet {
				for pid := range st.rolePerms[roleID] {
					if p, ok := st.permissions[pid]; ok {
						grantingByPerm[p.Name] = append(grantingByPerm[p.Name], g.Name)
					}
				}
			}
		}
		for name, groups := range grantingByPerm {
			sort.Strings(groups)
			grantingByPerm[name] = groups
		}
	}

	auditRows := make([]MyPermissionAuditRow, 0, len(permCatalog))
	for _, p := range permCatalog {
		row := MyPermissionAuditRow{
			Permission: p.Name,
			Granted:    isSuperuser || effectivePerms[p.Name],
		}
		if isSuperuser {
			row.GrantingGroups = nil
		} else {
			row.GrantingGroups = grantingByPerm[p.Name]
			if row.GrantingGroups == nil {
				row.GrantingGroups = []string{}
			}
		}
		auditRows = append(auditRows, row)
	}

	return groupNames, roleNames, isSuperuser, auditRows, nil
}

func (m *Memory) ListUserGroups(_ context.Context) ([]UserGroupRow, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	out := make([]UserGroupRow, 0, len(st.groups))
	for id := range st.groups {
		g, err := st.userGroupRowLocked(id)
		if err != nil {
			return nil, err
		}
		if g != nil {
			out = append(out, *g)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (m *Memory) GetUserGroupByName(_ context.Context, name string) (*UserGroupRow, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	id, ok := st.groupByName[name]
	if !ok {
		return nil, nil
	}
	return st.userGroupRowLocked(id)
}

func (m *Memory) GetUserGroupByID(_ context.Context, id int) (*UserGroupRow, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	return st.userGroupRowLocked(id)
}

func (m *Memory) CreateUserGroup(_ context.Context, name string) (int, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	if _, ok := st.groupByName[name]; ok {
		return 0, fmt.Errorf("create user group: name already exists")
	}
	now := iamNow()
	id := st.nextGroupID
	st.nextGroupID++
	st.groups[id] = memoryGroup{ID: id, Name: name, CreatedAt: now, UpdatedAt: now}
	st.groupByName[name] = id
	st.memberships[id] = map[int]struct{}{}
	return id, nil
}

func (m *Memory) DeleteUserGroup(_ context.Context, id int) error {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	g, ok := st.groups[id]
	if !ok {
		return sql.ErrNoRows
	}
	if isSystemUserGroup(g.Name) {
		return fmt.Errorf("cannot delete system group %q", g.Name)
	}
	delete(st.groups, id)
	delete(st.groupByName, g.Name)
	delete(st.memberships, id)
	delete(st.groupRoles, id)
	return nil
}

func (m *Memory) ListUserGroupMemberIDs(_ context.Context, groupID int) ([]int, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	members := st.memberships[groupID]
	out := make([]int, 0, len(members))
	for userID := range members {
		out = append(out, userID)
	}
	sort.Ints(out)
	return out, nil
}

func (m *Memory) ListUserGroupIDsForUser(_ context.Context, userID int) ([]int, error) {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	out := make([]int, 0)
	for groupID, members := range st.memberships {
		if _, ok := members[userID]; ok {
			out = append(out, groupID)
		}
	}
	sort.Ints(out)
	return out, nil
}

func (m *Memory) SetUserGroupMembers(_ context.Context, groupID int, userIDs []int) error {
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	if _, ok := st.groups[groupID]; !ok {
		return sql.ErrNoRows
	}
	next := map[int]struct{}{}
	for _, userID := range userIDs {
		next[userID] = struct{}{}
	}
	st.memberships[groupID] = next
	return st.ensureSuperuserCoverageLocked()
}
