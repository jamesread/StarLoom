# Architecture

StarApp follows the [jwr-soa-2.0](https://github.com/jamesread) layout.

```
StarApp/
├── database/          # sql-migrate trees (sqlite/)
├── protocol/          # Protobuf + buf code generation
├── service/           # Go ConnectRPC backend
├── frontend/          # Vue + Vite + PicoCrank SPA
├── integration-tests/ # Mocha + Selenium end-to-end tests
└── docs/              # MkDocs documentation
```

## Request flow

1. The SPA calls `Init` on load for shell metadata (title, footer, version).
2. Authenticated routes use `GetStatus` for session and RBAC; it repeats the
   same footer fields as Init.
3. Domain RPCs use Connect JSON-over-HTTP under `/api` (Vite proxy in dev).

## Storage

The `Store` interface in `service/internal/store` abstracts persistence.
SQLite is the supported driver for self-hosted deployments.

Migrations live in `database/sqlite/migrations/` and are applied with
sql-migrate **before** the service starts (container entrypoint or `make migrate`).
At startup the service asserts that `config.RequiredMigration`
(`6.domain-rbac.sql`) is present in the `migrations` table.

## Observability

Prometheus metrics are exposed at `/metrics` (e.g. `starapp_init_total`).
