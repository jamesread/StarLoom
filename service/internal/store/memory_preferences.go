package store

import "context"

func (m *Memory) GetUserPreferences(ctx context.Context, userID int) (*UserPreferencesRow, error) {
	_ = ctx
	if prefs, ok := m.userPrefs[userID]; ok {
		copy := prefs
		return &copy, nil
	}
	return DefaultUserPreferences(userID), nil
}

func (m *Memory) SaveUserPreferences(ctx context.Context, userID int, language string, sidebarEnabled bool) error {
	_ = ctx
	if m.userPrefs == nil {
		m.userPrefs = map[int]UserPreferencesRow{}
	}
	m.userPrefs[userID] = UserPreferencesRow{
		UserAccountID:  userID,
		Language:       language,
		SidebarEnabled: sidebarEnabled,
	}
	return nil
}
