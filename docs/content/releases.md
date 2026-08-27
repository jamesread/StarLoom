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

The **Release Pipeline** workflow (`.github/workflows/release.yml`) runs `make
build`, `make test`, and `make lint` on every push and pull request. On pushes to
`main`, semantic-release creates a GitHub Release when commits warrant a version
bump, then GoReleaser builds and publishes multi-arch container images to
`ghcr.io/jamesread/starloom`.

### Container images

After a release:

```bash
docker pull ghcr.io/jamesread/starloom:latest
# or pin a version:
docker pull ghcr.io/jamesread/starloom:1.2.3
```

The image serves the built SPA from `/usr/share/starapp/webui`, stores SQLite
data under `/config` (volume), and runs migrations on startup.

### GitHub Actions permissions

The **release** job sets `packages: write` so the built-in `GITHUB_TOKEN` can
push container images to GHCR. No extra repository secrets are required.

If releases fail to create tags (e.g. branch protection on `main`), allow GitHub
Actions to bypass protection or adjust protection rules for the release bot.

## GoReleaser

Container images are built via `.goreleaser.yaml` (multi-arch `linux/amd64` and
`linux/arm64`, manifest tags `{{version}}` and `latest`). GoReleaser is invoked
from semantic-release's `@semantic-release/exec` `publishCmd` after the GitHub
Release is created.

## First release

If the repository already shipped manually, seed the last tag to match history:

```bash
git tag 0.1.0 <commit-sha>
git push origin 0.1.0
```

Use `v0.1.0` only if you change `tagFormat` to `'v${version}'` in `.releaserc.yaml`.
