package store

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// ListChores returns a family's chores in display order.
func (m *Memory) ListChores(ctx context.Context, familyID int, starChartID int, includeInactive bool) ([]ChoreWithAssignments, error) {
	st := m.choreState()
	st.mu.Lock()
	defer st.mu.Unlock()
	m.ensureDefaultStarChartLocked(familyID)
	out := filterMemoryChores(st.chores, familyID, starChartID, includeInactive)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Chore.SortOrder != out[j].Chore.SortOrder {
			return out[i].Chore.SortOrder < out[j].Chore.SortOrder
		}
		return out[i].Chore.Title < out[j].Chore.Title
	})
	return out, nil
}

func filterMemoryChores(chores map[int]ChoreWithAssignments, familyID, starChartID int, includeInactive bool) []ChoreWithAssignments {
	var out []ChoreWithAssignments
	for _, cw := range chores {
		if memoryChoreMatches(cw, familyID, starChartID, includeInactive) {
			out = append(out, cw)
		}
	}
	return out
}

func memoryChoreMatches(cw ChoreWithAssignments, familyID, starChartID int, includeInactive bool) bool {
	if cw.Chore.FamilyID != familyID {
		return false
	}
	return chartAndActiveMatch(cw, starChartID, includeInactive)
}

func chartAndActiveMatch(cw ChoreWithAssignments, starChartID int, includeInactive bool) bool {
	if starChartID > 0 && cw.Chore.StarChartID != starChartID {
		return false
	}
	if !includeInactive && !cw.Chore.Active {
		return false
	}
	return true
}

func (m *Memory) GetChoreByID(_ context.Context, id int) (*ChoreWithAssignments, error) {
	st := m.choreState()
	st.mu.Lock()
	defer st.mu.Unlock()
	if cw, ok := st.chores[id]; ok {
		cp := cw
		return &cp, nil
	}
	return nil, nil
}

// CreateChore appends a chore to the end of the order.
func (m *Memory) CreateChore(_ context.Context, familyID, starChartID int, title string, starReward, weekdayMask int, childMemberIDs []int) (int, error) {
	st := m.choreState()
	st.mu.Lock()
	defer st.mu.Unlock()
	if starChartID == 0 {
		starChartID = m.ensureDefaultStarChartLocked(familyID)
	}
	if weekdayMask == 0 {
		weekdayMask = 127
	}
	return m.insertMemoryChore(st, familyID, starChartID, title, starReward, weekdayMask, childMemberIDs), nil
}

func (m *Memory) insertMemoryChore(st *memoryChore, familyID, starChartID int, title string, starReward, weekdayMask int, childMemberIDs []int) int {
	st.nextChoreID++
	id := st.nextChoreID
	c := ChoreRow{
		ID: id, FamilyID: familyID, StarChartID: starChartID, Title: title, StarReward: starReward,
		WeekdayMask: weekdayMask, Active: true, SortOrder: maxChoreSort(st.chores, familyID) + 1, CreatedAt: familyNow(),
	}
	st.chores[id] = ChoreWithAssignments{Chore: c, Assignments: newMemoryAssignments(st, id, childMemberIDs)}
	return id
}

func maxChoreSort(chores map[int]ChoreWithAssignments, familyID int) int {
	maxSort := 0
	for _, cw := range chores {
		if cw.Chore.FamilyID == familyID && cw.Chore.SortOrder > maxSort {
			maxSort = cw.Chore.SortOrder
		}
	}
	return maxSort
}

func newMemoryAssignments(st *memoryChore, choreID int, childMemberIDs []int) []ChoreAssignmentRow {
	var assigns []ChoreAssignmentRow
	for _, childID := range childMemberIDs {
		st.nextAssignmentID++
		assigns = append(assigns, ChoreAssignmentRow{ID: st.nextAssignmentID, ChoreID: choreID, ChildMemberID: childID})
	}
	return assigns
}

func (m *Memory) UpdateChore(_ context.Context, id int, starChartID int, title string, starReward, weekdayMask int, active bool, childMemberIDs []int) error {
	st := m.choreState()
	st.mu.Lock()
	defer st.mu.Unlock()
	cw, ok := st.chores[id]
	if !ok {
		return fmt.Errorf("chore not found")
	}
	if weekdayMask == 0 {
		weekdayMask = 127
	}
	if starChartID == 0 {
		starChartID = m.ensureDefaultStarChartLocked(cw.Chore.FamilyID)
	}
	cw.Chore.Title = title
	cw.Chore.StarChartID = starChartID
	cw.Chore.StarReward = starReward
	cw.Chore.WeekdayMask = weekdayMask
	cw.Chore.Active = active
	var assigns []ChoreAssignmentRow
	for _, childID := range childMemberIDs {
		st.nextAssignmentID++
		assigns = append(assigns, ChoreAssignmentRow{ID: st.nextAssignmentID, ChoreID: id, ChildMemberID: childID})
	}
	cw.Assignments = assigns
	st.chores[id] = cw
	return nil
}

// ReorderChores numbers the given chores from one.
func (m *Memory) ReorderChores(_ context.Context, familyID int, choreIDs []int) error {
	st := m.choreState()
	st.mu.Lock()
	defer st.mu.Unlock()
	for i, id := range choreIDs {
		cw, ok := st.chores[id]
		if !ok || cw.Chore.FamilyID != familyID {
			continue
		}
		cw.Chore.SortOrder = i + 1
		st.chores[id] = cw
	}
	return nil
}

func (m *Memory) DeactivateChore(_ context.Context, id int) error {
	st := m.choreState()
	st.mu.Lock()
	defer st.mu.Unlock()
	cw, ok := st.chores[id]
	if !ok {
		return fmt.Errorf("chore not found")
	}
	cw.Chore.Active = false
	st.chores[id] = cw
	return nil
}

func (m *Memory) ListChorePauses(_ context.Context, familyID int) ([]ChorePauseRow, error) {
	st := m.choreState()
	st.mu.Lock()
	defer st.mu.Unlock()
	var out []ChorePauseRow
	for _, p := range st.pauses {
		if p.FamilyID == familyID {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartDate > out[j].StartDate })
	return out, nil
}

func (m *Memory) CreateChorePause(_ context.Context, familyID int, startDate, endDate, reason string) (int, error) {
	st := m.choreState()
	st.mu.Lock()
	defer st.mu.Unlock()
	st.nextPauseID++
	id := st.nextPauseID
	st.pauses[id] = ChorePauseRow{
		ID: id, FamilyID: familyID, StartDate: startDate, EndDate: endDate,
		Reason: reason, CreatedAt: familyNow(),
	}
	return id, nil
}

func (m *Memory) DeleteChorePause(_ context.Context, id int) error {
	st := m.choreState()
	st.mu.Lock()
	defer st.mu.Unlock()
	delete(st.pauses, id)
	return nil
}

func (m *Memory) IsDatePaused(_ context.Context, familyID int, date string) (bool, error) {
	st := m.choreState()
	st.mu.Lock()
	defer st.mu.Unlock()
	for _, p := range st.pauses {
		if p.FamilyID == familyID && date >= p.StartDate && date <= p.EndDate {
			return true, nil
		}
	}
	return false, nil
}

func (m *Memory) GetAssignment(_ context.Context, choreID, childMemberID int) (*ChoreAssignmentRow, error) {
	st := m.choreState()
	st.mu.Lock()
	defer st.mu.Unlock()
	cw, ok := st.chores[choreID]
	if !ok {
		return nil, nil
	}
	for _, a := range cw.Assignments {
		if a.ChildMemberID == childMemberID {
			cp := a
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *Memory) GetCompletion(_ context.Context, assignmentID int, date string) (*ChoreCompletionRow, error) {
	st := m.choreState()
	st.mu.Lock()
	defer st.mu.Unlock()
	for _, c := range st.completions {
		if c.AssignmentID == assignmentID && c.CompletionDate == date {
			cp := c
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *Memory) InsertChoreCompletion(_ context.Context, assignmentID int, date string, ledgerEntryID int) (int, error) {
	st := m.choreState()
	st.mu.Lock()
	defer st.mu.Unlock()
	st.nextCompletionID++
	id := st.nextCompletionID
	le := ledgerEntryID
	st.completions[id] = ChoreCompletionRow{
		ID: id, AssignmentID: assignmentID, CompletionDate: date,
		LedgerEntryID: &le, CreatedAt: familyNow(),
	}
	return id, nil
}

func (m *Memory) DeleteChoreCompletion(_ context.Context, id int) error {
	st := m.choreState()
	st.mu.Lock()
	defer st.mu.Unlock()
	delete(st.completions, id)
	return nil
}

func (m *Memory) ListCompletionsForWeek(_ context.Context, familyID int, weekStart, weekEnd string) ([]WeeklyChartCompletion, error) {
	st := m.choreState()
	st.mu.Lock()
	defer st.mu.Unlock()
	assignChore, assignChild, choreReward := memoryAssignmentMaps(st.chores, familyID)
	return filterWeekCompletions(st.completions, assignChore, assignChild, choreReward, weekStart, weekEnd), nil
}

func memoryAssignmentMaps(chores map[int]ChoreWithAssignments, familyID int) (map[int]int, map[int]int, map[int]int) {
	assignChore := map[int]int{}
	assignChild := map[int]int{}
	choreReward := map[int]int{}
	for id, cw := range chores {
		if cw.Chore.FamilyID != familyID {
			continue
		}
		choreReward[id] = cw.Chore.StarReward
		for _, a := range cw.Assignments {
			assignChore[a.ID] = a.ChoreID
			assignChild[a.ID] = a.ChildMemberID
		}
	}
	return assignChore, assignChild, choreReward
}

func filterWeekCompletions(completions map[int]ChoreCompletionRow, assignChore, assignChild, choreReward map[int]int, weekStart, weekEnd string) []WeeklyChartCompletion {
	var out []WeeklyChartCompletion
	for _, c := range completions {
		choreID, ok := assignChore[c.AssignmentID]
		if !ok || c.CompletionDate < weekStart || c.CompletionDate > weekEnd {
			continue
		}
		out = append(out, WeeklyChartCompletion{
			AssignmentID: c.AssignmentID, ChoreID: choreID,
			ChildMemberID: assignChild[c.AssignmentID], CompletionDate: c.CompletionDate,
			StarsEarned: choreReward[choreID], LedgerEntryID: c.LedgerEntryID,
		})
	}
	return out
}

func (m *Memory) ListChoreLedgerEntryIDs(_ context.Context, familyID int) ([]int, error) {
	st := m.choreState()
	st.mu.Lock()
	defer st.mu.Unlock()
	return ledgerIDsForFamilyCompletions(st.chores, st.completions, familyID), nil
}

func ledgerIDsForFamilyCompletions(chores map[int]ChoreWithAssignments, completions map[int]ChoreCompletionRow, familyID int) []int {
	assignFamily := familyAssignmentSet(chores, familyID)
	var out []int
	for _, c := range completions {
		if assignFamily[c.AssignmentID] && c.LedgerEntryID != nil {
			out = append(out, *c.LedgerEntryID)
		}
	}
	return out
}

func familyAssignmentSet(chores map[int]ChoreWithAssignments, familyID int) map[int]bool {
	assignFamily := map[int]bool{}
	for _, cw := range chores {
		if cw.Chore.FamilyID == familyID {
			for _, a := range cw.Assignments {
				assignFamily[a.ID] = true
			}
		}
	}
	return assignFamily
}

func (m *Memory) ListBonusStarsForWeek(ctx context.Context, familyID int, weekStart, weekEnd string, choreLedgerIDs []int) (map[int]map[string]int, error) {
	choreSet := intSet(choreLedgerIDs)
	fam := m.familyState()
	fam.mu.Lock()
	defer fam.mu.Unlock()
	return accumulateBonusStars(fam.ledger, familyID, weekStart, weekEnd, choreSet), nil
}

func intSet(ids []int) map[int]bool {
	out := map[int]bool{}
	for _, id := range ids {
		out[id] = true
	}
	return out
}

func accumulateBonusStars(ledger map[int]StarLedgerRow, familyID int, weekStart, weekEnd string, choreSet map[int]bool) map[int]map[string]int {
	out := map[int]map[string]int{}
	for _, e := range ledger {
		d, ok := bonusStarDate(e, familyID, weekStart, weekEnd, choreSet)
		if !ok {
			continue
		}
		if out[e.ChildMemberID] == nil {
			out[e.ChildMemberID] = map[string]int{}
		}
		out[e.ChildMemberID][d] += e.Amount
	}
	return out
}

func bonusStarDate(e StarLedgerRow, familyID int, weekStart, weekEnd string, choreSet map[int]bool) (string, bool) {
	if !isBonusAwardCandidate(e, familyID, choreSet) {
		return "", false
	}
	d := ledgerDate(e.CreatedAt)
	if d < weekStart || d > weekEnd {
		return "", false
	}
	return d, true
}

func isBonusAwardCandidate(e StarLedgerRow, familyID int, choreSet map[int]bool) bool {
	if e.FamilyID != familyID || e.EntryType != LedgerTypeAward {
		return false
	}
	if strings.HasPrefix(e.Note, "Chore:") || choreSet[e.ID] {
		return false
	}
	return true
}

func ledgerDate(createdAt string) string {
	if len(createdAt) >= 10 {
		return createdAt[:10]
	}
	return createdAt
}

type memoryChore struct {
	chores           map[int]ChoreWithAssignments
	pauses           map[int]ChorePauseRow
	completions      map[int]ChoreCompletionRow
	nextChoreID      int
	nextAssignmentID int
	nextPauseID      int
	nextCompletionID int
	mu               sync.Mutex
}

func (m *Memory) choreState() *memoryChore {
	if m.chores == nil {
		m.chores = &memoryChore{
			chores:      map[int]ChoreWithAssignments{},
			pauses:      map[int]ChorePauseRow{},
			completions: map[int]ChoreCompletionRow{},
		}
	}
	return m.chores
}
