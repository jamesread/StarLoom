package store

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

type memoryFamily struct {
	mu sync.Mutex

	nextFamilyID     int
	families         map[int]FamilyRow
	nextMemberID     int
	members          map[int]FamilyMemberRow
	nextLedgerID     int
	ledger           map[int]StarLedgerRow
	nextRewardID     int
	rewards          map[int]RewardRow
	nextRedemptionID int
	redemptions      map[int]RedemptionRow
}

func (m *Memory) familyState() *memoryFamily {
	if m.family == nil {
		m.family = &memoryFamily{
			families:    map[int]FamilyRow{},
			members:     map[int]FamilyMemberRow{},
			ledger:      map[int]StarLedgerRow{},
			rewards:     map[int]RewardRow{},
			redemptions: map[int]RedemptionRow{},
		}
	}
	return m.family
}

func familyNow() string {
	return time.Now().UTC().Format("2006-01-02 15:04:05")
}

func (m *Memory) CountFamilies(ctx context.Context) (int, error) {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	return len(st.families), nil
}

func (m *Memory) GetFamilyByID(_ context.Context, id int) (*FamilyRow, error) {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	if f, ok := st.families[id]; ok {
		cp := f
		return &cp, nil
	}
	return nil, nil
}

func (m *Memory) GetFirstFamily(_ context.Context) (*FamilyRow, error) {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	if len(st.families) == 0 {
		return nil, nil
	}
	ids := make([]int, 0, len(st.families))
	for id := range st.families {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	f := st.families[ids[0]]
	return &f, nil
}

func (m *Memory) CreateFamily(_ context.Context, name string) (int, error) {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	st.nextFamilyID++
	id := st.nextFamilyID
	st.families[id] = FamilyRow{ID: id, Name: name, CreatedAt: familyNow()}
	return id, nil
}

func (m *Memory) UpdateFamilyName(_ context.Context, id int, name string) error {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	f, ok := st.families[id]
	if !ok {
		return fmt.Errorf("family not found")
	}
	f.Name = name
	st.families[id] = f
	return nil
}

func (m *Memory) enrichMember(st *memoryFamily, mrow FamilyMemberRow) FamilyMemberRow {
	if mrow.UserAccountID != nil {
		if u, err := m.GetUserByID(context.Background(), *mrow.UserAccountID); err == nil && u != nil {
			mrow.Username = u.Username
		}
	}
	return mrow
}

func (m *Memory) GetMemberByID(_ context.Context, id int) (*FamilyMemberRow, error) {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	if row, ok := st.members[id]; ok {
		cp := m.enrichMember(st, row)
		return &cp, nil
	}
	return nil, nil
}

func (m *Memory) GetMemberByAccountID(_ context.Context, accountID int) (*FamilyMemberRow, error) {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	for _, row := range st.members {
		if row.UserAccountID != nil && *row.UserAccountID == accountID {
			cp := m.enrichMember(st, row)
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *Memory) ListMembersByFamily(_ context.Context, familyID int) ([]FamilyMemberRow, error) {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	out := []FamilyMemberRow{}
	for _, row := range st.members {
		if row.FamilyID == familyID {
			out = append(out, m.enrichMember(st, row))
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Role != out[j].Role {
			return out[i].Role < out[j].Role
		}
		return out[i].DisplayName < out[j].DisplayName
	})
	return out, nil
}

func (m *Memory) CreateMember(_ context.Context, familyID int, displayName, role string, accountID *int, starColor string) (int, error) {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	st.nextMemberID++
	id := st.nextMemberID
	st.members[id] = FamilyMemberRow{
		ID: id, FamilyID: familyID, UserAccountID: accountID,
		DisplayName: displayName, Role: role, StarColor: starColor, CreatedAt: familyNow(),
	}
	return id, nil
}

func (m *Memory) UpdateMember(_ context.Context, id int, displayName, starColor string) error {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	row, ok := st.members[id]
	if !ok {
		return fmt.Errorf("member not found")
	}
	row.DisplayName = displayName
	row.StarColor = starColor
	st.members[id] = row
	return nil
}

func (m *Memory) DeleteMember(_ context.Context, id int) error {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	row, ok := st.members[id]
	if !ok || row.Role != MemberRoleChild {
		return fmt.Errorf("member not found")
	}
	delete(st.members, id)
	return nil
}

func (m *Memory) SetMemberAvatarPath(_ context.Context, id int, path string) error {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	row, ok := st.members[id]
	if !ok {
		return fmt.Errorf("member not found")
	}
	row.AvatarPath = path
	st.members[id] = row
	return nil
}

func (m *Memory) GetMemberBalance(_ context.Context, memberID int) (int, error) {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	sum := 0
	for _, e := range st.ledger {
		if e.ChildMemberID == memberID {
			sum += e.Amount
		}
	}
	return sum, nil
}

func (m *Memory) ListLedger(_ context.Context, memberID, limit int) ([]StarLedgerRow, error) {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	if limit <= 0 {
		limit = 50
	}
	rows := []StarLedgerRow{}
	for _, e := range st.ledger {
		if e.ChildMemberID == memberID {
			rows = append(rows, e)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].CreatedAt != rows[j].CreatedAt {
			return rows[i].CreatedAt > rows[j].CreatedAt
		}
		return rows[i].ID > rows[j].ID
	})
	if len(rows) > limit {
		rows = rows[:limit]
	}
	return rows, nil
}

func (m *Memory) GetLastAward(_ context.Context, memberID int) (*StarLedgerRow, error) {
	rows, err := m.ListLedger(context.Background(), memberID, 100)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		if rows[i].EntryType == LedgerTypeAward {
			cp := rows[i]
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *Memory) InsertLedgerEntry(_ context.Context, row StarLedgerRow) (int, error) {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	st.nextLedgerID++
	id := st.nextLedgerID
	row.ID = id
	if row.CreatedAt == "" {
		row.CreatedAt = familyNow()
	}
	st.ledger[id] = row
	return id, nil
}

func (m *Memory) ListRewards(_ context.Context, familyID int, includeInactive bool) ([]RewardRow, error) {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	out := []RewardRow{}
	for _, r := range st.rewards {
		if r.FamilyID != familyID {
			continue
		}
		if !includeInactive && !r.Active {
			continue
		}
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CostStars != out[j].CostStars {
			return out[i].CostStars < out[j].CostStars
		}
		return out[i].Title < out[j].Title
	})
	return out, nil
}

func (m *Memory) GetRewardByID(_ context.Context, id int) (*RewardRow, error) {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	if r, ok := st.rewards[id]; ok {
		cp := r
		return &cp, nil
	}
	return nil, nil
}

func (m *Memory) CreateReward(_ context.Context, familyID int, title, description string, costStars int, approvalRequired bool) (int, error) {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	st.nextRewardID++
	id := st.nextRewardID
	st.rewards[id] = RewardRow{
		ID: id, FamilyID: familyID, Title: title, Description: description,
		CostStars: costStars, Active: true, ApprovalRequired: approvalRequired,
	}
	return id, nil
}

func (m *Memory) UpdateReward(_ context.Context, id int, title, description string, costStars int, active, approvalRequired bool) error {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	r, ok := st.rewards[id]
	if !ok {
		return fmt.Errorf("reward not found")
	}
	r.Title = title
	r.Description = description
	r.CostStars = costStars
	r.Active = active
	r.ApprovalRequired = approvalRequired
	st.rewards[id] = r
	return nil
}

func (m *Memory) DeactivateReward(_ context.Context, id int) error {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	r, ok := st.rewards[id]
	if !ok {
		return fmt.Errorf("reward not found")
	}
	r.Active = false
	st.rewards[id] = r
	return nil
}

func (m *Memory) ListRedemptions(_ context.Context, familyID int, status string) ([]RedemptionRow, error) {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	out := []RedemptionRow{}
	for _, r := range st.redemptions {
		if familyID > 0 && r.FamilyID != familyID {
			continue
		}
		if status != "" && r.Status != status {
			continue
		}
		if rw, ok := st.rewards[r.RewardID]; ok {
			r.RewardTitle = rw.Title
		}
		if mrow, ok := st.members[r.ChildMemberID]; ok {
			r.ChildDisplayName = mrow.DisplayName
		}
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt != out[j].CreatedAt {
			return out[i].CreatedAt > out[j].CreatedAt
		}
		return out[i].ID > out[j].ID
	})
	return out, nil
}

func (m *Memory) GetRedemptionByID(_ context.Context, id int) (*RedemptionRow, error) {
	rows, err := m.ListRedemptions(context.Background(), 0, "")
	if err != nil {
		return nil, err
	}
	for i := range rows {
		if rows[i].ID == id {
			cp := rows[i]
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *Memory) CreateRedemption(_ context.Context, familyID, childMemberID, rewardID, starsSpent int, status string, ledgerEntryID *int) (int, error) {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	st.nextRedemptionID++
	id := st.nextRedemptionID
	st.redemptions[id] = RedemptionRow{
		ID: id, FamilyID: familyID, ChildMemberID: childMemberID, RewardID: rewardID,
		StarsSpent: starsSpent, Status: status, LedgerEntryID: ledgerEntryID, CreatedAt: familyNow(),
	}
	return id, nil
}

func (m *Memory) ResolveRedemption(_ context.Context, id int, status string, resolvedByMemberID int, ledgerEntryID *int) error {
	st := m.familyState()
	st.mu.Lock()
	defer st.mu.Unlock()
	r, ok := st.redemptions[id]
	if !ok {
		return fmt.Errorf("redemption not found")
	}
	r.Status = status
	r.ResolvedAt = familyNow()
	r.ResolvedByMemberID = &resolvedByMemberID
	r.LedgerEntryID = ledgerEntryID
	st.redemptions[id] = r
	return nil
}

func (m *Memory) CountPendingRedemptions(_ context.Context, familyID int) (int, error) {
	rows, err := m.ListRedemptions(context.Background(), familyID, RedemptionPending)
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

func (m *Memory) EnsureUserInGroup(ctx context.Context, userID int, groupName string) error {
	g, err := m.GetUserGroupByName(ctx, groupName)
	if err != nil || g == nil {
		return fmt.Errorf("group %q not found", groupName)
	}
	ids, err := m.ListUserGroupMemberIDs(ctx, g.ID)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if id == userID {
			return nil
		}
	}
	ids = append(ids, userID)
	return m.SetUserGroupMembers(ctx, g.ID, ids)
}
