# Webhooks

Webhooks send an HTTP POST to a URL you choose when something happens in
the family — a star award, a redemption request, or an approval. Use them
to ping a chat bot, a home-automation hook, or another service you run.

You do not need webhooks for everyday awarding and redeeming.

## Create a target

1. Sign in as an administrator.
2. Open **Control Panel → Webhooks**.
3. Add a webhook: URL, optional secret, and which events to send.
4. Leave it enabled.

The secret is write-only. StarLoom will not show it again after you save.

## Events

| Event | When it fires |
|-------|----------------|
| `stars.awarded` | A parent awards stars |
| `redemption.requested` | Someone requests a reward (often pending approval) |
| `redemption.resolved` | A parent approves or rejects a request |
| `webhooks.test` | You click test on the webhook page |

## What is sent

Each delivery is JSON:

```json
{
  "event": "stars.awarded",
  "timestamp": "2026-08-27T20:00:00Z"
}
```

Headers:

- `Content-Type: application/json`
- `X-StarApp-Event: <event name>`
- `X-StarApp-Signature: sha256=<hex>`

The signature is HMAC-SHA256 of the raw body using your secret. Check it
before you trust the payload:

```
expected = HMAC_SHA256(secret, raw_body)
header == "sha256=" + hex(expected)
```

Deliveries use a short timeout. A failing hook does not block awarding or
redeeming in the app.
