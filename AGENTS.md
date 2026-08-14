# AGENTS.md

Watchtower is a Go 1.26.6 app (module `github.com/sidneyojr/watchtower`) that monitors running Docker containers and recreates them when their image updates. This fork is based on the actively maintained `nicholas-fedor/watchtower` (upstream `containrrr/watchtower` is archived). Branches: `master` = stable/release, `develop` = integration.

## Git Flow
- `develop` is the integration branch — all work (features, fixes, changelog) lands here first.
- `master` only receives pull requests whose head branch is `develop` (enforced by `.github/workflows/enforce-pr-source.yaml` + the "Protect master (Git Flow)" ruleset). Direct pushes to `master` are rejected by the ruleset (PR required, no force-push/deletion).
- Flow: commit on `develop` → push → open PR `develop`→`master` → enforcer check passes → squash-merge. **Do NOT pass `--delete-branch` when merging** — the head branch IS `develop`, so the remote branch would be deleted (recreate with `git push origin develop`). Before opening a new PR, sync `develop` with `master` first (squash merges diverge the SHAs even when content is identical): `git checkout develop && git pull --ff-only origin develop && git merge --ff-only origin/master && git push origin develop`. `--ff-only` fails on the squash divergence — then use a normal `git merge origin/master -m "chore: sync develop with master after squash merge"` (conventional message required; `merge:` type is rejected by the hook).
- Releases: tag on `master` (`v*`) triggers `release-stable.yaml` → goreleaser publishes multi-arch images to `ghcr.io/sidneyojr/watchtower` with tags `latest`, `<major>`, `<major>.<minor>`, `<major>.<minor>.<patch>` (the exact image tag has NO `v` prefix). Remember to sync `develop` after tagging.
- `gh` may resolve the wrong repo when both `origin` and `upstream` remotes exist — run `gh repo set-default sidneyojr/watchtower`. Push over HTTPS fails without credentials; `origin` uses SSH.

## Fork-specific changes
This fork diverges from upstream `nicholas-fedor/watchtower` in these intentional ways. Do NOT regress them:
- **Image-name inspect fallback** (`pkg/container/container_source.go`, commit `e133b641`): when inspecting a container's image by ID fails (`No such image`), fall back to inspecting the configured image name. Rootless Docker with BuildKit removes the previous image when a tag is rebuilt even while a container still runs from it, so the ID-based inspect fails and watchtower would otherwise skip the update.
- **Lifecycle hooks gated on `IsStale()`** (`internal/actions/update.go`, commit `bebff44b`): pre/post-update hooks run ONLY for containers with a new image (`IsStale`), not for linked-only restarts (which restart without a new image). This restores the original documented gate removed by upstream PR #908; `scripts/lifecycle-tests.sh` case 3 depends on it.
- **`update-go-docs.yaml` pins `go-version: "1.26.6"`** (commit `93347bab`): the `go-proxy-pull-action` default `go-version: 1.26` takes precedence over `go-version-file`, so the refresh would fail with an older toolchain. Do not switch back to `go-version-file`.
- **`release-stable.yaml` dispatches `publish-docs.yaml`** with the tag (commit `17b081f0`): the `release` event does NOT fire for releases created with `GITHUB_TOKEN` (goreleaser), so the docs workflow never ran on release. The `publish-docs` job dispatches it with `VERSION=<tag>` and `ALIASES=latest`.
- **`publish-docs.yaml` calls `mike set-default`** (commit `c5ef59db`): `mike deploy` writes versioned subdirs but not the root `index.html` redirect; without `set-default` the docs root 404s.

## Layout
- `cmd/` — cobra CLI. Entry: `main.go` → `cmd.Execute()`. The whole app flow (flag wiring, scheduler, HTTP API, update loop) is in `cmd/root.go` (`PreRun`/`Run`); `cmd/notify-upgrade.go` is a subcommand.
- `internal/actions/` — core update orchestration (`Update`, `CheckForSanity`, `CheckForMultipleWatchtowerInstances`). The heart of the tool.
- `internal/api/` — HTTP API (fiber v3): auth, lifecycle, routes, handlers, swagger.
- `internal/config/` — viper-based configuration (api, client, docker, filter, lifecycle, ...).
- `internal/logging/` — zerolog logging setup.
- `internal/meta/` — `Version`, set ONLY via build-time ldflags (defaults to `v0.0.0-unknown`).
- `internal/scheduling/`, `internal/metrics/`, `internal/flags/`, `internal/util/` — helpers.
- `pkg/container/` — Docker client wrapper (`container.Client`) and container metadata/recreation.
- `pkg/registry/` — registry auth/digest/manifest/trust handling.
- `pkg/notifications/` — shoutrrr-based notifiers.
- `pkg/session/` — `Progress` accumulates per-run results into a `types.Report`.
- `pkg/types/` — shared cross-package interfaces (`Container`, `Filter`, `UpdateParams`, `Report`).
- `pkg/compose`, `pkg/filters`, `pkg/lifecycle`, `pkg/sorter` — self-explanatory helpers.
- `tools/tplprev/` — SEPARATE Go module `github.com/sidneyojr/tplprev` (CLI + WASM template preview). Excluded from root `go build ./...`.
- `build/` — build tooling: `docker/` (Dockerfiles), `goreleaser/` (`stable.yaml`, `nightly.yaml`), `golangci-lint/`, `mockery/`, `mkdocs/`.
- `test/` — test data and shell scripts (`test/scripts/test_tls_integration.sh`).

## Build
- `make build` compiles the binary to `bin/watchtower` (a plain `go build ./...` reports `v0.0.0-unknown` in `internal/meta.Version`).
- Releases use goreleaser v2 (`build/goreleaser/stable.yaml`, `nightly.yaml`), injecting `-X github.com/sidneyojr/watchtower/internal/meta.Version={{ .Version }}` and building multi-arch Docker images published ONLY to `ghcr.io/sidneyojr/watchtower` (no Docker Hub).
- `make run`, `make install`, `make docker-build`, `make release` also available. Toolchain lives in `~/.local/go` and `~/go/bin` (golangci-lint, mockery, goreleaser); export `PATH="/home/sidney/.local/go/bin:/home/sidney/go/bin:$PATH"` in new shells.

## Tests
- `go test ./...` (or `make test`) needs no Docker daemon — tests are ginkgo/gomega BDD suites (`*_suite_test.go`) with hand-written mocks (`internal/actions/mocks`, `pkg/container/mocks`, ...).
- Real integration tests are shell scripts needing a running Docker daemon and a locally built binary:
  - `scripts/lifecycle-tests.sh` — lifecycle hooks (slow: builds/restarts containers).
  - `scripts/dependency-test.sh` — depends-on/linked-container restart ordering.
  - `scripts/contnet-tests.sh` — container networking; requires `VPN_SERVICE_PROVIDER`, `OPENVPN_USER`, `OPENVPN_PASSWORD` env vars.
- `tools/tplprev` has its own tests: `make tplprev-test`.

## Lint
- `make lint` runs golangci-lint with `--fix --config build/golangci-lint/golangci-lint.yaml ./...`; `make vet` runs `go vet ./...`; `make fmt` runs `golangci-lint fmt`. There are `tplprev-lint`/`tplprev-vet`/`tplprev-fmt` equivalents.

## Conventional Commits
- The `commit-msg` hook in `.githooks/` enforces Conventional Commits: `^(feat|fix|docs|style|refactor|test|chore|build|ci|perf|revert)(\([a-zA-Z0-9_.-]+\))?(!)?:\s.*$`. It is versioned; enable with `git config core.hooksPath .githooks`. The default merge message is rejected — always commit merges with a conventional message.

## Docs / template preview
- Docs are mkdocs (`build/mkdocs/mkdocs.yaml`, sources in `docs/`); deps in `build/mkdocs/docs-requirements.txt`. Site publishes to `https://sidneyojr.github.io/watchtower/` from `master`.
- Pages is enabled on the repo serving the `gh-pages` branch (legacy build). The mike deploy writes versioned subdirs (`dev/`, `latest/`, `v1.21.1/`); `mike set-default` creates the root `index.html` redirect to the first alias — without it the root 404s. Root works only when a versioned release has been published.
- `publish-docs.yaml` triggers on: `push` to `master` touching docs (deploys `dev`), `release` published, and `workflow_dispatch` with optional `VERSION`/`ALIASES` inputs (deploys that version + alias). The `release` event does NOT fire for releases created with `GITHUB_TOKEN` (goreleaser), so `release-stable.yaml` has a `publish-docs` job that dispatches the workflow with the tag name. To (re)publish the docs for an existing release manually: `gh workflow run publish-docs.yaml --ref master -f VERSION=vX.Y.Z -f ALIASES=latest`.
- `docs/template-preview.md` embeds `tplprev.wasm`. Regenerate with `scripts/build-tplprev.sh` (compiles `./tools/tplprev` for `GOOS=js GOARCH=wasm` and copies `wasm_exec.js` from GOROOT). Outputs are gitignored.
- `tools/tplprev/` is split by build tags (`//go:build !wasm` vs `//go:build wasm`), so the wasm target is excluded from `go build ./...`.
- `docs/README.md` documents the docs-site workflow (mike versioning, publish-docs.yaml).

## CI/CD
- GitHub Actions in `.github/workflows/` (no CircleCI): `test.yaml` and `lint-go.yaml` run on PRs touching Go code; `build.yaml` is a reusable workflow called by releases; `release-stable.yaml` on `v*` tags; `release-nightly.yaml` on a daily cron; `publish-docs.yaml` on `master`; plus `security.yaml`, `scorecard.yml`, `update-changelog.yaml`, `clean-cache.yaml`, `update-go-docs.yaml`, `lint-gh.yaml`, and `enforce-pr-source.yaml` (blocks PRs into `master` whose head is not `develop`; this check is required by the "Protect master (Git Flow)" ruleset). Branch filters target `master`/`develop`.
- `update-changelog.yaml` now targets `develop` (trigger on `develop` push, PR base `develop`, merged via `gh pr merge --squash` without `--auto`, since `--auto` requires branch protection that `develop` doesn't have). The changelog reaches `master` via the `develop`→`master` PR.