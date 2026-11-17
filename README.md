<div align="center">
  <img alt="StarLoom logo" src="logo.svg" width="128" />
  <h1>StarLoom</h1>

  Family star rewards — parents award stars for good behavior; children redeem them for privileges like screen time.

[![Maturity Badge](https://img.shields.io/badge/maturity-Beta-yellow)](#none)
[![Discord](https://img.shields.io/discord/846737624960860180?label=Discord%20Server)](https://discord.gg/jhYWWpNJ3v)

</div>

## Screenshots

### Login

Sign-in screen in the Space theme, where parents and children enter the family star rewards app.

<p align="center">
<img alt="Login page in the Space theme" src="var/marketing/screenshots/login-space.png" />
</p>

### Parent home

Family overview for a privileged parent in the Egypt theme, with children, stars and quick actions in one place.

<p align="center">
<img alt="Parent home in the Egypt theme" src="var/marketing/screenshots/parent-home-egypt.png" />
</p>

### Child home

A child's personal home in the Aztecs theme, showing their star balance and rewards they can redeem.

<p align="center">
<img alt="Child home in the Aztecs theme" src="var/marketing/screenshots/child-home-aztecs.png" />
</p>

### Weekly star chart

The weekly star chart in the Ancient Greece theme, where parents track chores and award stars day by day.

<p align="center">
<img alt="Star chart in the Ancient Greece theme" src="var/marketing/screenshots/star-chart-greece.png" />
</p>

### Rewards

Rewards and redemption requests in the Space theme with dark mode, so parents can approve privileges earned with stars.

<p align="center">
<img alt="Rewards admin in the Space theme, dark mode" src="var/marketing/screenshots/rewards-space-dark.png" />
</p>

### People

Family people admin in the Aztecs theme with dark mode, listing members and their roles.

<p align="center">
<img alt="People admin in the Aztecs theme, dark mode" src="var/marketing/screenshots/people-aztecs-dark.png" />
</p>

### Control Panel

The system Control Panel in the Catppuccin theme, the hub for IAM, settings and other privileged tools.

<p align="center">
<img alt="Control Panel in the Catppuccin theme" src="var/marketing/screenshots/control-panel-catppuccin.png" />
</p>

### User preferences

User preferences in the Egypt theme, where each person sets language, sidebar and theme options.

<p align="center">
<img alt="User preferences in the Egypt theme" src="var/marketing/screenshots/preferences-egypt.png" />
</p>

Screenshots can be regenerated with `make screenshots` (Vite SPA must be running at https://localhost:5173).

## Documentation

- [User manual](docs/content/index.md) — install, setup, and daily use
- [Product specification](docs/SPEC.md)
- [AGENTS.md](AGENTS.md) — AI agent integration (MCP, OpenAPI, llms.txt)

## Linux container

StarLoom is published as a multi-arch image on the GitHub Container Registry (`linux/amd64` and `linux/arm64`).

```bash
docker run -d \
  --name starapp \
  --restart unless-stopped \
  -p 8080:8080 \
  -v starapp-config:/config \
  -e STARAPP_SECURE_COOKIES=false \
  ghcr.io/jamesread/starloom:latest
```

`podman run` accepts the same arguments. Open http://localhost:8080 and sign in as **admin** / **admin**. Change that password immediately.

SQLite data and `config.yaml` persist on the `/config` volume. Omit `STARAPP_SECURE_COOKIES=false` when you terminate TLS at a reverse proxy. Pin a release with `ghcr.io/jamesread/starloom:1.2.3` instead of `latest` if you prefer.

## Releases

Container images are published to `ghcr.io/jamesread/starloom` on each release.
See [Install](docs/content/install.md) for pull tags and updates.

## Status

StarLoom is a self-hosted family ledger for stars, chores, and rewards. See the
[user manual](docs/content/index.md) to get started, or the
[product spec](docs/SPEC.md) for the full roadmap.
