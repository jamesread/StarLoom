# StarApp

Family star rewards — parents award stars for good behavior; children redeem them for privileges like screen time.

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

jwr-soa-2.0 skeleton in place: Connect RPC backend with Init, Vue + PicoCrank SPA, integration-test harness, and MkDocs. Domain features (ledger, rewards) are specified but not yet implemented.
