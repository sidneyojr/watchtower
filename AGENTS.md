# AGENTS.md

Watchtower is a Go 1.26.6 app (module `github.com/sidneyojr/watchtower`) that monitors running Docker containers and recreates them when their image updates. This fork is based on the actively maintained `nicholas-fedor/watchtower` (upstream `containrrr/watchtower` is archived). Branches: `master` = stable/release, `develop` = integration. Push both.

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
- `docs/template-preview.md` embeds `tplprev.wasm`. Regenerate with `scripts/build-tplprev.sh` (compiles `./tools/tplprev` for `GOOS=js GOARCH=wasm` and copies `wasm_exec.js` from GOROOT). Outputs are gitignored.
- `tools/tplprev/` is split by build tags (`//go:build !wasm` vs `//go:build wasm`), so the wasm target is excluded from `go build ./...`.
- `docs/README.md` documents the docs-site workflow (mike versioning, publish-docs.yaml).

## CI/CD
- GitHub Actions in `.github/workflows/` (no CircleCI): `test.yaml` and `lint-go.yaml` run on PRs touching Go code; `build.yaml` is a reusable workflow called by releases; `release-stable.yaml` on `v*` tags; `release-nightly.yaml` on a daily cron; `publish-docs.yaml` on `master`; plus `security.yaml`, `scorecard.yml`, `update-changelog.yaml`, `clean-cache.yaml`, `update-go-docs.yaml`, `lint-gh.yaml`. Branch filters target `master`/`develop`.