# Configuration

Configuration is split between **YAML/env bootstrap** and **database cvars**
(live runtime settings edited on the Settings page).

## YAML bootstrap

Loaded from the config directory passed to `-configdir` (default:
`~/.config/starapp`).

Copy `service/config.yaml.example` to `~/.config/starapp/config.yaml`:

```yaml
http_addr: ":8080"
db_driver: sqlite
db_path: starapp.db
show_footer: true   # seeds show_footer cvar on first insert only
# webui_dir: /path/to/frontend/dist
```

YAML covers values needed before the database is available (listen address,
database path). Feature toggles and site title live in **cvars** after the first
startup upsert.

## Runtime settings (cvars)

Admins edit cvars at **Control Panel › Settings** (`/control-panel/settings`). Values are stored in the
`cvars` table. The catalog (keys, types, titles, defaults) lives in
`service/internal/cvar/`.

On each startup, after migrations, the service upserts missing keys and refreshes
**metadata only** (title, description, category, ordinal). Admin-chosen values
are never overwritten.

Saving Settings reloads **Init** in the SPA (no process restart).

### Page footer

Footer visibility, version text, and update hints come from **Init** (and
**GetStatus**, which mirrors the same shell fields). The SPA renders
`AppFooter` under `#content` after `<main>`; femtocrank styles apply to
`footer` and `footer span` chips.

When `show_version_number` is off, Init clears `current_version` and
`available_version` on the server so the client cannot display them. When
`show_footer` is off, the footer is omitted entirely.

| Key | Type | Default | Category | Effect |
|-----|------|---------|----------|--------|
| `site_title` | string | StarApp | Site | Header and browser tab title |
| `show_footer` | bool | on | Site | Footer visibility (Init) |
| `show_version_number` | bool | on | Site | Version text in footer (Init) |
| `show_new_versions` | bool | on | Site | Update hint when a newer release exists (Init) |
| `default_award_stars` | int | 1 | Features | Default stars when awarding (1–100) |
| `enable_redemption_approval` | bool | on | Features | New rewards require approval by default (Init `features`) |

Parent authentication uses httpauthshim, session cookies, and group-based RBAC.
See [Authentication](authentication.md).

## Database migrations

Schema changes are versioned SQL files under `database/sqlite/migrations/`,
applied with [sql-migrate](https://github.com/rubenv/sql-migrate).

The service binary expects migration **`7.chores.sql`** (`config.RequiredMigration`).

On startup the service **applies pending migrations automatically** (using
`database/sqlite/migrations` relative to the working directory, or
`STARAPP_MIGRATIONS_DIR`). You can still run them manually:

### Development (manual)

```bash
export DB_PATH="$HOME/.config/starapp/starapp.db"
make migrate
```

Check status:

```bash
DB_PATH="$HOME/.config/starapp/starapp.db" make migrate-status
```

### Containers

The Docker image runs `sql-migrate up` in the entrypoint before starting the
service. Set `DB_PATH` to an absolute SQLite file path.

| Variable | Purpose |
|----------|---------|
| `DB_PATH` | SQLite database file (required for migrations) |
| `DB_DRIVER` | `sqlite` (default) |

## Environment

| Variable | Effect |
|----------|--------|
| `PORT` | When set, golure picks listen address ($PORT → 8080 → free 8000–8999). |

## Required migration

Current required migration id: **`2.webhook-targets-events.sql`** — webhook targets
and event subscriptions (after `0.base.sql` and `1.cvars.sql`).
