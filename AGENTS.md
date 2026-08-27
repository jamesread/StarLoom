# AGENTS.md

Guide for AI agents integrating with StarApp.

## Project overview

StarApp is a family star-rewards app: parents award stars for good behavior; children redeem them for privileges. The stack follows jwr-soa-2.0:

- **Backend**: Go, Connect RPC over HTTP
- **Frontend**: Vue 3, Vite, PicoCrank
- **Database**: SQLite with sql-migrate

Domain flows (families, stars, rewards, redemptions) are implemented. Settings (cvars) and outbound webhooks are available via the API and MCP.

## Discovery endpoints

All integration surfaces share the app origin:

| Path | Content-Type | Purpose |
|------|--------------|---------|
| `/llms.txt` | text/plain | llmstxt.org index |
| `/openapi` | application/json | OpenAPI 3.1 spec (Connect RPC) |
| `/mcp` | MCP Streamable HTTP | MCP tools for agents |
| `/api/*` | Connect JSON | Primary RPC API |

## Authentication

Sign in with username/password (session cookie `starapp-sid`) or a Bearer API key. MCP requires Bearer API keys only.

Default first user on empty database: **admin** / **admin** — change immediately in production.

Read-only API keys cannot use mutating MCP tools or RPCs that change state.

## MCP server

- **Endpoint**: `GET/POST /mcp` on the same base URL as the app (e.g. `http://localhost:8080/mcp`).
- **Transport**: MCP Streamable HTTP.
- **Auth**: Same as Connect API (currently open).

### Cursor configuration (HTTP)

```json
{
  "mcpServers": {
    "starapp": {
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

Add an `Authorization` header when parent auth is enabled.

### MCP tools

| Tool | Description | Parameters |
|------|-------------|------------|
| `starapp_ping` | Health check via Init | — |
| `starapp_init` | SPA bootstrap metadata | — |
| `starapp_list_cvars` | List configuration variables | — |
| `starapp_update_cvar` | Update a cvar | `key` (required), `value_int`, `value_string` |
| `starapp_list_webhooks` | List webhook targets | — |
| `starapp_create_webhook` | Create webhook | `url`, `secret`, `events` (required); `enabled` |
| `starapp_update_webhook` | Update webhook | `id` (required); `url`, `secret`, `events`, `enabled` |
| `starapp_delete_webhook` | Delete webhook | `id` (required) |

Webhook `events` use comma-separated names: `stars.awarded`, `redemption.requested`, `redemption.resolved`.

## Connect RPC API

Mount prefix: `/api` (also available without prefix for direct backend access).

Service: `starapp.api.v1.StarAppService`

| RPC | Purpose |
|-----|---------|
| `Init` | SPA shell metadata, features, webhook event catalog |
| `ListCvars` | List cvars |
| `UpdateCvar` | Update a cvar |
| `ListWebhooks` | List webhook targets |
| `CreateWebhook` | Create webhook |
| `UpdateWebhook` | Update webhook |
| `DeleteWebhook` | Delete webhook |

Example Init call:

```bash
curl -sS -X POST http://localhost:8080/api/starapp.api.v1.StarAppService/Init \
  -H 'Content-Type: application/json' \
  -d '{}'
```

For full request/response schemas, use `/openapi` or the protobuf definitions under `protocol/starapp/api/v1/`.

## Development

```bash
make              # build protocol, frontend, service
make -C service run
make -C frontend run   # HTTPS :5173, proxies /api
```

Regenerate OpenAPI after proto changes:

```bash
make -C protocol
```

## Security notes

- Webhook secrets are never returned by list/get RPCs or MCP list tools.
- Run StarApp on trusted networks until parent auth ships.
- Prefer HTTPS in production; terminate TLS at your reverse proxy.

## Documentation

- [Product spec](docs/SPEC.md)
- [Webhooks](docs/content/webhooks.md)
- [Configuration](docs/content/configuration.md)
