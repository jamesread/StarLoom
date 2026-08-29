# Releases

StarApp uses [semantic-release](https://semantic-release.gitbook.io/) on pushes to
`main`. Versions and GitHub Releases are created automatically from
[Conventional Commits](https://www.conventionalcommits.org/).

Configuration: `.releaserc.yaml` (tag format `1.2.3`, no `v` prefix).

## Conventional commits

| Prefix | Release bump |
|--------|----------------|
| `feat:` | Minor |
| `fix:`, `perf:` | Patch |
| `docs:`, `style:`, `refactor:`, `test:`, `build:`, `chore:`, `ci:` | No release |
| `feat!:` or `BREAKING CHANGE:` in the footer | Major |

Examples:

```
feat(rewards): add star award RPC

fix(webhooks): reject invalid event names (#12)
```

Link issues and pull requests with closing keywords in commit messages or PR
descriptions (`fixes #123`, `closes #456`) so `@semantic-release/github` can
comment on success.

## CI

The **Release Pipeline** workflow (`.github/workflows/release.yml`) runs on every
push and pull request (except tag-only pushes, which reuse the tested main
commit):

- `make build`
- `make test` (unit tests)
- `make lint`
- `make integration-test`

On pushes to `main`, when commits warrant a version bump:

1. **release** — semantic-release creates a single GitHub Release and git tag.
2. **publish-image** — native `linux/amd64` and `linux/arm64` images are built in
   parallel on `ubuntu-latest` and `ubuntu-24.04-arm`, then pushed to GHCR.
3. **publish-manifest** — combines the per-arch images into `{{version}}` and
   `latest` manifest tags (runs after both arch jobs succeed).

When there is nothing to release, image jobs are skipped.

### Container images

After a release:

```bash
docker pull ghcr.io/jamesread/starloom:latest
# or pin a version:
docker pull ghcr.io/jamesread/starloom:1.2.3
```

Per-arch tags are also published (`1.2.3-amd64`, `1.2.3-arm64`).

The image serves the built SPA from `/usr/share/starapp/webui`, stores SQLite
data under `/config` (volume), and runs migrations on startup.

### GitHub Actions permissions

The **publish-image** and **publish-manifest** jobs set `packages: write` so the
built-in `GITHUB_TOKEN` can push container images to GHCR. No extra repository
secrets are required.

If releases fail to create tags (e.g. branch protection on `main`), allow GitHub
Actions to bypass protection or adjust protection rules for the release bot.

## First release

If the repository already shipped manually, seed the last tag to match history:

```bash
git tag 0.1.0 <commit-sha>
git push origin 0.1.0
```

Use `v0.1.0` only if you change `tagFormat` to `'v${version}'` in `.releaserc.yaml`.
