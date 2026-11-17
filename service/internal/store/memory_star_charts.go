package store

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

type memoryStarChart struct {
	starCharts      map[int]StarChartRow
	nextStarChartID int
	mu              sync.Mutex
}

func (m *Memory) starChartState() *memoryStarChart {
	if m.starCharts == nil {
		m.starCharts = &memoryStarChart{starCharts: map[int]StarChartRow{}}
	}
	return m.starCharts
}

func (m *Memory) ensureDefaultStarChartLocked(familyID int) int {
	st := m.starChartState()
	for _, sc := range st.starCharts {
		if sc.FamilyID == familyID {
			return sc.ID
		}
	}
	st.nextStarChartID++
	id := st.nextStarChartID
	st.starCharts[id] = StarChartRow{
		ID: id, FamilyID: familyID, Name: "Star Chart", SortOrder: 0, Active: true, CreatedAt: familyNow(),
	}
	return id
}

func (m *Memory) ListStarCharts(_ context.Context, familyID int, includeInactive bool) ([]StarChartRow, error) {
	st := m.starChartState()
	st.mu.Lock()
	defer st.mu.Unlock()
	m.ensureDefaultStarChartLocked(familyID)
	out := filterStarCharts(st.starCharts, familyID, includeInactive)
	sort.Slice(out, func(i, j int) bool {
		if out[i].SortOrder != out[j].SortOrder {
			return out[i].SortOrder < out[j].SortOrder
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

func filterStarCharts(charts map[int]StarChartRow, familyID int, includeInactive bool) []StarChartRow {
	var out []StarChartRow
	for _, sc := range charts {
		if starChartMatches(sc, familyID, includeInactive) {
			out = append(out, sc)
		}
	}
	return out
}

func starChartMatches(sc StarChartRow, familyID int, includeInactive bool) bool {
	if sc.FamilyID != familyID {
		return false
	}
	if !includeInactive && !sc.Active {
		return false
	}
	return true
}

func (m *Memory) GetStarChartByID(_ context.Context, id int) (*StarChartRow, error) {
	st := m.starChartState()
	st.mu.Lock()
	defer st.mu.Unlock()
	if sc, ok := st.starCharts[id]; ok {
		cp := sc
		return &cp, nil
	}
	return nil, nil
}

func (m *Memory) GetDefaultStarChartID(_ context.Context, familyID int) (int, error) {
	st := m.starChartState()
	st.mu.Lock()
	defer st.mu.Unlock()
	id := m.ensureDefaultStarChartLocked(familyID)
	return id, nil
}

func (m *Memory) CreateStarChart(_ context.Context, familyID int, name string, sortOrder int) (int, error) {
	st := m.starChartState()
	st.mu.Lock()
	defer st.mu.Unlock()
	st.nextStarChartID++
	id := st.nextStarChartID
	st.starCharts[id] = StarChartRow{
		ID: id, FamilyID: familyID, Name: name, SortOrder: sortOrder, Active: true, CreatedAt: familyNow(),
	}
	return id, nil
}

func (m *Memory) UpdateStarChart(_ context.Context, id int, name string, sortOrder int, active bool) error {
	st := m.starChartState()
	st.mu.Lock()
	defer st.mu.Unlock()
	sc, ok := st.starCharts[id]
	if !ok {
		return fmt.Errorf("star chart not found")
	}
	sc.Name = name
	sc.SortOrder = sortOrder
	sc.Active = active
	st.starCharts[id] = sc
	return nil
}

func (m *Memory) DeleteStarChart(_ context.Context, id int) error {
	st := m.starChartState()
	st.mu.Lock()
	defer st.mu.Unlock()
	delete(st.starCharts, id)
	return nil
}

func (m *Memory) CountChoresForStarChart(_ context.Context, starChartID int) (int, error) {
	ch := m.choreState()
	ch.mu.Lock()
	defer ch.mu.Unlock()
	count := 0
	for _, cw := range ch.chores {
		if cw.Chore.StarChartID == starChartID {
			count++
		}
	}
	return count, nil
}

func (m *Memory) CountChoresForStarChartAndMember(_ context.Context, starChartID, memberID int) (int, error) {
	ch := m.choreState()
	ch.mu.Lock()
	defer ch.mu.Unlock()
	count := 0
	for _, cw := range ch.chores {
		if choreAssignedOnChart(cw, starChartID, memberID) {
			count++
		}
	}
	return count, nil
}

func choreAssignedOnChart(cw ChoreWithAssignments, starChartID, memberID int) bool {
	if cw.Chore.StarChartID != starChartID || !cw.Chore.Active {
		return false
	}
	for _, assign := range cw.Assignments {
		if assign.ChildMemberID == memberID {
			return true
		}
	}
	return false
}
