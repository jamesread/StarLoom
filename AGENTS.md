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

Webhook `events` use comma-separated names: `stars.awarded`, `redemption.requested`, `redemption.resolved`, `webhooks.test`.

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

## Frontend forms

Wrap every data-entry form in PicoCrank **`FormLayout`** with **`FormField`** children. Put primary and secondary actions in the `#actions` slot; use `@submit.prevent` on `FormLayout` and `type="submit"` on the primary button.

```vue
<FormLayout @submit.prevent="save">
  <FormField label="Title" for="example-title">
    <input id="example-title" v-model="title" type="text" required />
  </FormField>
  <FormField label="Options" component-has-label>
    <CheckGroup v-model="selected" :options="options" name="example-options" />
  </FormField>
  <template #actions>
    <button type="submit" class="good">Save</button>
    <button type="button" class="secondary" @click="cancel">Cancel</button>
  </template>
</FormLayout>
```

Use `component-has-label` on `FormField` when the control is not a single native input (e.g. `CheckGroup`, `RadioGroup`). Reference examples: `WebhookCreate.vue`, `ChangePassword.vue`, `SettingsAdmin.vue`.

The login screen uses PicoCrank’s `Login` component (`LoginForm.vue`); all other app forms follow the `FormLayout` pattern above.

## Buttons with icons

Use Femtocrank **`inline-icon`** on plain `<button>` elements (no PicoCrank `Button` component). Follow PicoCrank’s [buttons-with-icons](https://github.com/jamesread/picocrank/blob/main/docs/buttons-with-icons.md) guide (`node_modules/picocrank/docs/buttons-with-icons.md` in development):

1. Add **`inline-icon`** on the button; use **`neutral`** for secondary/toolbar actions, **`good`** for primary confirm.
2. Render **`HugeiconsIcon`** from `@hugeicons/vue` with icons from `@hugeicons/core-free-icons`.
3. Size icons at **`1em` × `1em`** with **`strokeWidth` `2.5`** for button-sized icons.
4. Set **`aria-hidden="true"`** on the icon when visible text follows; use **`aria-label`** on the button for icon-only buttons.

```vue
<script setup lang="ts">
import { HugeiconsIcon } from '@hugeicons/vue'
import { Refresh01Icon } from '@hugeicons/core-free-icons'

const iconStrokeWidth = 2.5
</script>

<template>
  <button type="button" class="inline-icon neutral" aria-label="Refresh">
    <HugeiconsIcon
      :icon="Refresh01Icon"
      width="1em"
      height="1em"
      :strokeWidth="iconStrokeWidth"
      aria-hidden="true"
    />
  </button>
  <button type="button" class="inline-icon neutral">
    <HugeiconsIcon
      :icon="ArrowLeft01Icon"
      width="1em"
      height="1em"
      :strokeWidth="iconStrokeWidth"
      aria-hidden="true"
    />
    <span>Previous</span>
  </button>
</template>
```

Section toolbars: put icon buttons in the Section **`#toolbar`** slot. Reference: `StarChart.vue`, `WebhooksAdmin.vue`, PicoCrank `vue/examples/ButtonsExample.vue`.

## Section toolbars

When adding page-level actions to a **`Section`** (refresh, create, navigation), put them in the section **title bar** via the **`#toolbar`** slot — not in a separate row inside the section body.

Toolbar rules:

- At most **one** **`good`** button per toolbar.
- That **`good`** button is always the **rightmost** control (primary action).
- Other actions use **`neutral`** (or icon-only with `aria-label`).

```vue
<Section title="Chores" :icon="Task01Icon">
  <template #toolbar>
    <button type="button" class="inline-icon neutral" aria-label="Refresh" @click="load">
      <HugeiconsIcon :icon="Refresh01Icon" width="1em" height="1em" :strokeWidth="2.5" aria-hidden="true" />
    </button>
    <button type="button" class="inline-icon neutral" @click="showPause = true">
      <HugeiconsIcon :icon="PauseIcon" width="1em" height="1em" :strokeWidth="2.5" aria-hidden="true" />
      <span>Pause chores</span>
    </button>
    <button type="button" class="inline-icon good" @click="showCreate = true">
      <HugeiconsIcon :icon="PlusSignIcon" width="1em" height="1em" :strokeWidth="2.5" aria-hidden="true" />
      <span>Add chore</span>
    </button>
  </template>
  <!-- section content -->
</Section>
```

Row-level destructive actions (e.g. Deactivate on a table row) may stay inline in the table. Reference: `ChoresAdmin.vue`, `StarChart.vue`, `WebhooksAdmin.vue`.

## Datatable row navigation

When a datatable row links to **edit**, **open**, or **detail**, always **navigate to a dedicated route** — never open an inline panel below the table, a second `Section` on the same page, or an edit dialog.

| Action | Pattern |
|--------|---------|
| Create | `<dialog>` on the list page, or a dedicated create route (e.g. `WebhookCreate.vue`) |
| Edit / open / detail | New route with `:id` param (e.g. `/family/chores/:id`, `/control-panel/webhooks/:id`) |
| Primary column | `RouterLink` to the edit/detail route |
| Actions column | `RouterLink` labeled Edit, or navigate on row click |

Edit pages include a **Back** link in the section `#toolbar` returning to the list. On save, `router.push` back to the list. Reference: `ChoreEdit.vue`, `WebhookEdit.vue`, `ChildDetail.vue`.

```vue
<RouterLink :to="{ name: 'familyChoreEdit', params: { id: row.id } }" class="title-link">
  <strong>{{ value }}</strong>
</RouterLink>
```

## Security notes

- Webhook secrets are never returned by list/get RPCs or MCP list tools.
- Run StarApp on trusted networks until parent auth ships.
- Prefer HTTPS in production; terminate TLS at your reverse proxy.

## Documentation

- [Product spec](docs/SPEC.md)
- [Webhooks](docs/content/webhooks.md)
- [Configuration](docs/content/configuration.md)
