<div align="center">
  <img alt="StarApp logo" src="logo.svg" width="128" />
  <h1>StarApp</h1>

  Family star rewards — parents award stars for good behavior; children redeem them for privileges like screen time.

[![Maturity Badge](https://img.shields.io/badge/maturity-Beta-yellow)](#none)
[![Discord](https://img.shields.io/discord/846737624960860180?label=Discord%20Server)](https://discord.gg/jhYWWpNJ3v)

</div>

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
