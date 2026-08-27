package store

import (
	"context"
	"sort"
)

func (m *Memory) ListCvars(_ context.Context) ([]CvarRow, error) {
	out := make([]CvarRow, 0, len(m.cvars))
	for _, row := range m.cvars {
		out = append(out, row)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Ordinal != out[j].Ordinal {
			return out[i].Ordinal < out[j].Ordinal
		}
		return out[i].Key < out[j].Key
	})
	return out, nil
}

func (m *Memory) FindCvar(_ context.Context, key string) (*CvarRow, error) {
	row, ok := m.cvars[key]
	if !ok {
		return nil, nil
	}
	copy := row
	return &copy, nil
}

func (m *Memory) InsertCvarIfMissing(_ context.Context, row CvarRow) error {
	if existing, ok := m.cvars[row.Key]; ok {
		existing.Title = row.Title
		existing.Description = row.Description
		existing.Category = row.Category
		existing.Ordinal = row.Ordinal
		m.cvars[row.Key] = existing
		return nil
	}
	m.cvars[row.Key] = row
	return nil
}

func (m *Memory) UpdateCvar(_ context.Context, key string, valueInt int, valueString string) error {
	row, ok := m.cvars[key]
	if !ok {
		return errCvarNotFound
	}
	row.ValueInt = valueInt
	row.ValueString = valueString
	m.cvars[key] = row
	return nil
}
