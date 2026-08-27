package store

type UserPreferencesRow struct {
	UserAccountID      int
	Language           string
	SidebarEnabled     bool
	ThemeToggleEnabled bool
}

func DefaultUserPreferences(userID int) *UserPreferencesRow {
	return &UserPreferencesRow{
		UserAccountID:      userID,
		Language:           "",
		SidebarEnabled:     true,
		ThemeToggleEnabled: false,
	}
}
