package store

import (
	"context"
	"database/sql"
	"fmt"
)

func (s *SQLite) GetUserPreferences(ctx context.Context, userID int) (*UserPreferencesRow, error) {
	var lang string
	var sidebar, themeToggle int
	err := s.db.QueryRowContext(ctx, `
		SELECT language, sidebar_enabled, theme_toggle_enabled
		FROM user_preferences WHERE user_account_id = ?`, userID).
		Scan(&lang, &sidebar, &themeToggle)
	if err == sql.ErrNoRows {
		return DefaultUserPreferences(userID), nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user preferences: %w", err)
	}
	return &UserPreferencesRow{
		UserAccountID:      userID,
		Language:           lang,
		SidebarEnabled:     sidebar != 0,
		ThemeToggleEnabled: themeToggle != 0,
	}, nil
}

func (s *SQLite) SaveUserPreferences(ctx context.Context, userID int, language string, sidebarEnabled, themeToggleEnabled bool) error {
	sidebar := 0
	if sidebarEnabled {
		sidebar = 1
	}
	themeToggle := 0
	if themeToggleEnabled {
		themeToggle = 1
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO user_preferences (user_account_id, language, sidebar_enabled, theme_toggle_enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))
		ON CONFLICT(user_account_id) DO UPDATE SET
			language = excluded.language,
			sidebar_enabled = excluded.sidebar_enabled,
			theme_toggle_enabled = excluded.theme_toggle_enabled,
			updated_at = datetime('now')`,
		userID, language, sidebar, themeToggle)
	if err != nil {
		return fmt.Errorf("save user preferences: %w", err)
	}
	return nil
}
