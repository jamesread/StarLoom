package store

type UserPreferencesRow struct {
	Language       string
	UserAccountID  int
	SidebarEnabled bool
}

func DefaultUserPreferences(userID int) *UserPreferencesRow {
	return &UserPreferencesRow{
		UserAccountID:  userID,
		Language:       "",
		SidebarEnabled: true,
	}
}
