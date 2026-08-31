# Webhooks

StarApp can POST JSON to admin-configured URLs when domain events occur. Targets
and event subscriptions are stored in the database (`webhook_targets`,
`webhook_events`).

## Supported events

| Event | When fired (planned / future handlers) |
|-------|----------------------------------------|
| `stars.awarded` | Parent awards stars to a child |
| `redemption.requested` | Child requests a reward (especially pending approval) |
| `redemption.resolved` | Parent approves or rejects a redemption |
| `webhooks.test` | Manual test from the webhooks admin page |

The catalog is defined in code (`service/internal/webhook/`). The Settings UI
CheckGroup reads the list from Init / ListWebhooks.

## Payload shape

Every delivery includes:

```json
{
  "event": "stars.awarded",
  "timestamp": "2026-08-27T20:00:00Z"
}
```

Event-specific fields will be added alongside when domain handlers call the
dispatcher (e.g. `child`, `amount`, `reward`).

## Signing

Requests use:

- `Content-Type: application/json`
- `X-StarApp-Event: <event name>`
- `X-StarApp-Signature: sha256=<hex>`

The signature is HMAC-SHA256 of the **raw JSON body** using the target secret.

Verification sketch (pseudocode):

```
expected = HMAC_SHA256(secret, raw_body)
assert header == "sha256=" + hex(expected)
```

Secrets are **write-only** — never returned by List/Create/Update responses.

## Admin UI

- List: `/control-panel/webhooks` (PicoCrank Table)
- Create: `/control-panel/webhooks/create` (CheckGroup for events, boolean RadioGroup for enabled)

Dispatch uses a ~2s HTTP timeout and does not block user actions if delivery fails.
