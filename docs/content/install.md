# Install

The supported way to run StarApp is a **Linux container**. The image includes
the web app, the API, and SQLite. Migrations run automatically when the
container starts.

Images are published to the GitHub Container Registry for `linux/amd64` and
`linux/arm64`:

`ghcr.io/jamesread/starloom`

## Docker or Podman

```bash
docker run -d \
  --name starapp \
  --restart unless-stopped \
  -p 8080:8080 \
  -v starapp-config:/config \
  -e STARAPP_SECURE_COOKIES=false \
  ghcr.io/jamesread/starloom:latest
```

`podman run` accepts the same arguments.

Then open http://localhost:8080 and sign in as **admin** / **admin**. Change
that password immediately — see [First setup](setup.md).

`STARAPP_SECURE_COOKIES=false` lets the login cookie work over plain HTTP
(typical on a home LAN). Omit that variable when you terminate TLS at a
reverse proxy.

## Docker Compose

```yaml
services:
  starapp:
    image: ghcr.io/jamesread/starloom:latest
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - starapp-config:/config
    environment:
      STARAPP_SECURE_COOKIES: "false"

volumes:
  starapp-config:
```

Start it with `docker compose up -d`.

## What the volume stores

The `/config` volume holds:

- `config.yaml` — listen address and database path
- `starapp.db` — family data, users, stars, and settings
- uploaded avatars

Keep a copy of this volume if you care about history.

## Pin a version

Replace `latest` with a release tag when you want a fixed version:

```bash
ghcr.io/jamesread/starloom:1.2.3
```

## HTTPS

StarApp does not terminate TLS itself. Put a reverse proxy in front (Caddy,
nginx, Traefik) and talk to the container on port 8080.

When the browser uses HTTPS:

- Do **not** set `STARAPP_SECURE_COOKIES=false`
- Forward the usual headers (`Host`, `X-Forwarded-Proto`, `X-Forwarded-For`)

## Updates

Pull a newer image and recreate the container. The `/config` volume stays in
place; pending database migrations run on the next start.

```bash
docker pull ghcr.io/jamesread/starloom:latest
docker stop starapp && docker rm starapp
# run the same docker run (or compose up) as before
```

## Next

[Create your family and add people](setup.md).
