# StarApp

StarApp is a family rewards app. Parents award **stars** to children for good
behavior; children redeem stars for privileges such as screen time.

## Stack

- **Frontend** — Vue 3 + Vite + PicoCrank (`frontend/`)
- **Backend** — Go ConnectRPC service (`service/`)
- **Protocol** — Protocol Buffers + buf (`protocol/`)
- **Tests** — Mocha + Selenium integration tests (`integration-tests/`)

## Quick start

```sh
make -C service run
make -C frontend run
```

See [Architecture](architecture.md), [Configuration](configuration.md), and the
[Product spec](spec.md) for details.
