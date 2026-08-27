# Authentication

StarApp uses **group-based RBAC** with **httpauthshim** for HTTP authentication.

## Sign in

- **Username & password** — PicoCrank login form; session cookie `starapp-sid` (HttpOnly, SameSite=Strict).
- **Bearer API key** — `Authorization: Bearer <key>` for automation and MCP (no cookies on `/mcp`).
- **Optional SSO** — configure top-level `auth:` in `config.yaml` for JWT, trusted headers, or mTLS (see [httpauthshim](https://github.com/jamesread/httpauthshim)).

## First user

When the database has no accounts, startup creates **`admin`** / **`admin`**. Change this password immediately in production.

## Access model

Permissions are **never** attached directly to users:

1. Define **roles** and tick **permissions**.
2. Assign **roles to groups**.
3. Add **users to groups**.

System seeds:

| Group | Role |
|-------|------|
| Everyone | member (`app.access`) |
| Administrators | superuser (all permissions) |

## Development

| Variable | Effect |
|----------|--------|
| `STARAPP_DEV_DISABLE_AUTH=true` | Anonymous superuser (never in production) |
| `STARAPP_SECURE_COOKIES=false` | Allow session cookie over HTTP |

## Configuration example

```yaml
# ~/.config/starapp/config.yaml
auth: {}
```

Extend `auth:` with JWT, trusted headers, or mTLS per httpauthshim documentation. Strip inbound SSO headers at your reverse proxy unless you trust the source.

## User profile

Click the **username** in the header to open the [User Control Panel](user-profile.md): identity, preferences (language, sidebar, theme switcher), change password, API keys, My Permissions, and sign out.

Privileged system pages (IAM, Settings, webhooks) live under **Control Panel** in the sidebar. That hub is shown only to superusers and users with a matching permission (`users.view`, `usergroups.view`, `rbac.view`, or `system.settings`). It is not the User Control Panel.
