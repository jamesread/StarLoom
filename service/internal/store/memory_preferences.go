package store

import "context"

func (m *Memory) GetUserPreferences(ctx context.Context, userID int) (*UserPreferencesRow, error) {
	_ = ctx
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	if prefs, ok := st.userPrefs[userID]; ok {
		copy := prefs
		return &copy, nil
	}
	return DefaultUserPreferences(userID), nil
}

func (m *Memory) SaveUserPreferences(ctx context.Context, userID int, language string, sidebarEnabled, themeToggleEnabled bool) error {
	_ = ctx
	st := m.iamState()
	st.mu.Lock()
	defer st.mu.Unlock()
	if st.userPrefs == nil {
		st.userPrefs = map[int]UserPreferencesRow{}
	}
	st.userPrefs[userID] = UserPreferencesRow{
		UserAccountID:      userID,
		Language:           language,
		SidebarEnabled:     sidebarEnabled,
		ThemeToggleEnabled: themeToggleEnabled,
	}
	return nil
}
