# Configuration

Most day-to-day options live in the app under **Control Panel → Settings**.
A small YAML file and a few environment variables cover how the process
starts.

## Settings in the app

Open **Control Panel → Settings** while signed in as an administrator.
Changes apply immediately; you do not restart the container.

### Site

| Setting | What it does |
|---------|----------------|
| Site title | Name in the header and browser tab |
| Show footer | Show or hide the page footer |
| Show version number | Show the installed version in the footer |
| Show new versions | Offer a link when a newer release exists |

### Features

| Setting | What it does |
|---------|----------------|
| Default award stars | Prefill when a parent awards stars by hand (1–100) |
| Redemption approval by default | New rewards require parent approval before stars are spent |

### Theme

| Setting | What it does |
|---------|----------------|
| Color scheme switcher | Show an auto / light / dark control in the header |
| Theme name | Default look for the household (or the enforced look) |
| Theme control | **User preference** lets each person pick a theme. **System preference** forces the theme name above for everyone |

Built-in themes include Ancient Greece, Aztecs, Egypt, and Space. See
[Preferences and themes](preferences.md).

## File configuration

The container reads `/config/config.yaml` from the volume you mounted at
install time. A default file is created on first start.

```yaml
http_addr: ":8080"
db_driver: sqlite
db_path: starapp.db
show_footer: true
```

| Key | Purpose |
|-----|---------|
| `http_addr` | Listen address inside the container (default `:8080`) |
| `db_driver` | `sqlite` |
| `db_path` | SQLite file, relative to the config directory unless you use an absolute path |
| `show_footer` | Seeds the **Show footer** setting the first time only |
| `webui_dir` | Leave unset in the container; the image already serves the web UI |

`site_title` and other live options belong in Settings, not this file.
Restart the container after you edit YAML.

## Environment variables

Set these on the container (`docker run -e` or Compose `environment:`).

| Variable | When to use |
|----------|-------------|
| `STARAPP_SECURE_COOKIES` | Set to `false` only when the browser uses HTTP. Leave unset (secure cookies on) behind HTTPS |
| `PORT` | Alternate listen port if you do not use `http_addr` |
| `STARAPP_CONFIG_DIR` | Config directory (default `/config` in the image) |
| `DB_PATH` | Override the SQLite file path |
| `DB_DRIVER` | `sqlite` (default) |

Do not enable development-only flags such as disabled authentication on a
family install.

## Reverse proxy

Put Caddy, nginx, or Traefik in front of port 8080 when you want HTTPS or a
hostname such as `stars.home.example`.

- Proxy `/` to the container
- Forward `Host` and `X-Forwarded-Proto`
- Omit `STARAPP_SECURE_COOKIES=false` so the session cookie is marked Secure

## Backups

Copy the `/config` volume (or at least `starapp.db` and `avatars/`) on a
schedule. That is the household ledger, accounts, and pictures.

## Next

- [Accounts](accounts.md) — parents, children, and passwords
- [Preferences and themes](preferences.md)
- [Webhooks](webhooks.md) — optional HTTP notifications
