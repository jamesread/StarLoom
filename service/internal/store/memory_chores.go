package store

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

func (m *Memory) ListChores(ctx context.Context, familyID int, includeInactive bool) ([]ChoreWithAssignments, error) {
	st := m.choreState()
	st.mu.Lock()
	defer st.mu.Unlock()
	var out []ChoreWithAssignments
	for _, cw := range st.chores {
		if cw.Chore.FamilyID != familyID {
			continue
		}
		if !includeInactive && !cw.Chore.Active {
			continue
		}
		out = append(out, cw)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Chore.Title < out[j].Chore.Title })
	return out, nil
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

func (m *Memory) CreateChore(_ context.Context, familyID int, title string, starReward, weekdayMask int, childMemberIDs []int) (int, error) {
	st := m.choreState()
	st.mu.Lock()
	defer st.mu.Unlock()
	if weekdayMask == 0 {
		weekdayMask = 127
	}
	st.nextChoreID++
	id := st.nextChoreID
	c := ChoreRow{
		ID: id, FamilyID: familyID, Title: title, StarReward: starReward,
		WeekdayMask: weekdayMask, Active: true, CreatedAt: familyNow(),
	}
	var assigns []ChoreAssignmentRow
	for _, childID := range childMemberIDs {
		st.nextAssignmentID++
		assigns = append(assigns, ChoreAssignmentRow{ID: st.nextAssignmentID, ChoreID: id, ChildMemberID: childID})
	}
	st.chores[id] = ChoreWithAssignments{Chore: c, Assignments: assigns}
	return id, nil
}

func (m *Memory) UpdateChore(_ context.Context, id int, title string, starReward, weekdayMask int, active bool, childMemberIDs []int) error {
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
	cw.Chore.Title = title
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
	assignChore := map[int]int{}
	assignChild := map[int]int{}
	choreReward := map[int]int{}
	for id, cw := range st.chores {
		if cw.Chore.FamilyID != familyID {
			continue
		}
		choreReward[id] = cw.Chore.StarReward
		for _, a := range cw.Assignments {
			assignChore[a.ID] = a.ChoreID
			assignChild[a.ID] = a.ChildMemberID
		}
	}
	var out []WeeklyChartCompletion
	for _, c := range st.completions {
		choreID, ok := assignChore[c.AssignmentID]
		if !ok {
			continue
		}
		if c.CompletionDate < weekStart || c.CompletionDate > weekEnd {
			continue
		}
		out = append(out, WeeklyChartCompletion{
			AssignmentID: c.AssignmentID, ChoreID: choreID,
			ChildMemberID: assignChild[c.AssignmentID], CompletionDate: c.CompletionDate,
			StarsEarned: choreReward[choreID], LedgerEntryID: c.LedgerEntryID,
		})
	}
	return out, nil
}

func (m *Memory) ListChoreLedgerEntryIDs(_ context.Context, familyID int) ([]int, error) {
	st := m.choreState()
	st.mu.Lock()
	defer st.mu.Unlock()
	assignFamily := map[int]bool{}
	for _, cw := range st.chores {
		if cw.Chore.FamilyID == familyID {
			for _, a := range cw.Assignments {
				assignFamily[a.ID] = true
			}
		}
	}
	var out []int
	for _, c := range st.completions {
		if assignFamily[c.AssignmentID] && c.LedgerEntryID != nil {
			out = append(out, *c.LedgerEntryID)
		}
	}
	return out, nil
}

func (m *Memory) ListBonusStarsForWeek(ctx context.Context, familyID int, weekStart, weekEnd string, choreLedgerIDs []int) (map[string]int, error) {
	choreSet := map[int]bool{}
	for _, id := range choreLedgerIDs {
		choreSet[id] = true
	}
	fam := m.familyState()
	fam.mu.Lock()
	defer fam.mu.Unlock()
	out := map[string]int{}
	for _, e := range fam.ledger {
		if e.FamilyID != familyID || e.EntryType != LedgerTypeAward {
			continue
		}
		d := e.CreatedAt
		if len(d) >= 10 {
			d = d[:10]
		}
		if d < weekStart || d > weekEnd {
			continue
		}
		if choreSet[e.ID] {
			continue
		}
		out[d] += e.Amount
	}
	return out, nil
}

type memoryChore struct {
	mu sync.Mutex

	nextChoreID      int
	chores           map[int]ChoreWithAssignments
	nextAssignmentID int
	nextPauseID      int
	pauses           map[int]ChorePauseRow
	nextCompletionID int
	completions      map[int]ChoreCompletionRow
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
