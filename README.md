<div align="center">
  <img alt="StarApp logo" src="logo.svg" width="128" />
  <h1>StarApp</h1>

  Family star rewards — parents award stars for good behavior; children redeem them for privileges like screen time.

[![Maturity Badge](https://img.shields.io/badge/maturity-Beta-yellow)](#none)
[![Discord](https://img.shields.io/discord/846737624960860180?label=Discord%20Server)](https://discord.gg/jhYWWpNJ3v)

</div>

## Screenshots

Login (Space theme);

<p align="center">
<img alt="Login page in the Space theme" src="var/marketing/login-space.png" />
</p>

Parent home (Egypt theme) — family overview as a privileged user;

<p align="center">
<img alt="Parent home in the Egypt theme" src="var/marketing/parent-home-egypt.png" />
</p>

Child home (Aztecs theme) — personal stars and rewards;

<p align="center">
<img alt="Child home in the Aztecs theme" src="var/marketing/child-home-aztecs.png" />
</p>

Weekly star chart (Ancient Greece theme);

<p align="center">
<img alt="Star chart in the Ancient Greece theme" src="var/marketing/star-chart-greece.png" />
</p>

Rewards and redemption requests (Space theme, dark);

<p align="center">
<img alt="Rewards admin in the Space theme, dark mode" src="var/marketing/rewards-space-dark.png" />
</p>

People (Aztecs theme, dark);

<p align="center">
<img alt="People admin in the Aztecs theme, dark mode" src="var/marketing/people-aztecs-dark.png" />
</p>

Control Panel (Catppuccin theme);

<p align="center">
<img alt="Control Panel in the Catppuccin theme" src="var/marketing/control-panel-catppuccin.png" />
</p>

User preferences (Egypt theme);

<p align="center">
<img alt="User preferences in the Egypt theme" src="var/marketing/preferences-egypt.png" />
</p>

## Documentation

- [Product specification](docs/SPEC.md)
- [Developer docs](docs/content/index.md) (MkDocs)
- [AGENTS.md](AGENTS.md) — AI agent integration (MCP, OpenAPI, llms.txt)

## Run locally

```bash
# Terminal 1 — API (default :8080, or set PORT)
make -C service run

# Terminal 2 — frontend HTTPS on :5173 (proxies /api to backend)
make -C frontend run
```

Build everything: `make`

## Releases

Releases on `main` are automated via semantic-release from conventional commits.
See [Releases](docs/content/releases.md) for commit message format and CI behaviour.

## Status

StarApp implements the jwr-soa-2.0 stack: Connect RPC backend, Vue + PicoCrank SPA, family ledger, rewards, chores, webhooks, and integration tests. See the [product spec](docs/SPEC.md) for the full roadmap.
