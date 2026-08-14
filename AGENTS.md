# AGENTS.md

Watchtower is a Go 1.20 app (module `github.com/containrrr/watchtower`) that monitors running Docker containers and recreates them when their image updates. Upstream is unmaintained (see README banner). This checkout is on branch `master`, but the CI workflows in `.github/workflows/` trigger on `main` — don't "fix" that mismatch casually.

## Layout
- `cmd/` — cobra CLI. Entry: `main.go` → `cmd.Execute()`. The whole app flow (flag wiring, scheduler, HTTP API, update loop) is in `cmd/root.go` (`PreRun`/`Run`); `cmd/notify-upgrade.go` is a subcommand.
- `internal/actions/` — core update orchestration (`Update`, `CheckForSanity`, `CheckForMultipleWatchtowerInstances`). The heart of the tool.
- `internal/meta/` — `Version`, set ONLY via build-time ldflags (defaults to `v0.0.0-unknown`).
- `pkg/container/` — Docker client wrapper (`container.Client`) and container metadata/recreation.
- `pkg/registry/` — registry auth/digest/manifest/trust handling.
- `pkg/notifications/` — shoutrrr-based notifiers + Go template rendering (`preview/tplprev.go` renders a template against sample data).
- `pkg/api/` — HTTP API (`/v1/metrics`, `/v1/update`).
- `pkg/session/` — `Progress` accumulates per-run results into a `types.Report`.
- `pkg/types/` — shared cross-package interfaces (`Container`, `Filter`, `UpdateParams`, `Report`).
- `internal/flags`, `internal/util`, `pkg/filters`, `pkg/lifecycle`, `pkg/metrics`, `pkg/sorter` — self-explanatory helpers.

## Build
- `./build.sh` builds the `watchtower` binary and injects `meta.Version` via `-ldflags "-X github.com/containrrr/watchtower/internal/meta.Version=$(git describe --tags)"`. A plain `go build ./...` compiles but reports `v0.0.0-unknown`.
- Releases use goreleaser (`goreleaser.yml`), injecting the same ldflag and building multi-arch Docker images.

## Tests
- `go test ./...` needs no Docker daemon — tests are ginkgo/gomega BDD suites (`*_suite_test.go`) with hand-written mocks (`internal/actions/mocks`, `pkg/container/mocks`). CI runs `go test -v -coverprofile coverage.out -covermode atomic ./...`.
- Real integration tests are shell scripts needing a running Docker daemon and a locally built binary:
  - `scripts/lifecycle-tests.sh` — lifecycle hooks (slow: builds/restarts containers).
  - `scripts/dependency-test.sh` — depends-on/linked-container restart ordering.
  - `scripts/contnet-tests.sh` — container networking; requires `VPN_SERVICE_PROVIDER`, `OPENVPN_USER`, `OPENVPN_PASSWORD` env vars.
- `docker-compose.yml` brings up prometheus + grafana + dummy `parent`/`child` containers for manual testing (then run the binary with `--run-once` or `--interval`).

## Lint
- CI uses `staticcheck` (not golangci-lint). Its action pins `install-go: false` because the bundled toolchain is Go 1.17. Locally: `go vet ./...` or `staticcheck ./...`.

## Docs / template preview
- Docs are mkdocs (`mkdocs.yml`, sources in `docs/`); deps in `docs-requirements.txt`.
- `docs/template-preview.md` embeds `tplprev.wasm`. Regenerate with `scripts/build-tplprev.sh` (compiles `./tplprev` for `GOOS=js GOARCH=wasm` and copies `wasm_exec.js` from GOROOT). Outputs are gitignored.
- `tplprev/` is split by build tags (`//go:build !wasm` vs `//go:build wasm`), so the wasm target is excluded from `go build ./...`.