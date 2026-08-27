# User profile

Signed-in users open the **User Control Panel** by clicking their **username** in the header.

## Hub

| Area | Description |
|------|-------------|
| **Identity** | Username and account creation date from `GetStatus` |
| **Quick actions** | Preferences, change password, API keys, My Permissions |
| **Session** | Sign out (clears session cookie and reloads) |

Routes require authentication (`app.access`). This is self-service only — admins edit other users under **Control Panel › IAM › Users**.

## Preferences

Stored per user in `user_preferences` (not admin cvars):

| Setting | Default | Notes |
|---------|---------|-------|
| Language | Browser default (`""`) | Server validates against the i18n catalog |
| Sidebar | Enabled | Hides navigation sidebar when disabled |
| Theme switcher | Disabled | Shows or hides the header theme button |
| Color theme | Auto | Client-only via PicoCrank `useTheme` (localStorage) |

After login, the SPA loads preferences and applies sidebar and theme-switcher visibility immediately.

## Related RPCs

- `GetUserPreferences` / `SaveUserPreferences`
- `ChangePassword` (current + new password, min 8 characters)
- `Logout`

All act on the authenticated user only — no `user_id` on save.
