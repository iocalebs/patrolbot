# Repository Guidelines

## Project Structure & Module Organization
`main.go` is the CLI entrypoint. Most application code lives under `internal/`: `cli/` defines Cobra commands, `discord/` and `mediawiki/` wrap external APIs, `report/` builds report data and templates, and `config/` loads `config.yaml`. Root-level `templates/` contains report templates, `testdata/integration-tests/` stores `testscript` fixtures, `terraform/` holds deployment infrastructure, and `scripts/` contains small operational helpers.

## Build, Test, and Development Commands
Use `make` targets instead of ad hoc commands when possible:

- `make build` builds `./patrolbot` with `go build -trimpath`.
- `make test` runs the short test suite: `go test ./... -short`.
- `make testall` runs the full unit and integration suite.
- `make cover` generates merged coverage output in `coverage/`.
- `make lint` runs `golangci-lint run --fix ./...`.
- `make generate` runs `go generate ./...` and refreshes generated artifacts such as `config.schema.json`.
- `make serve` starts the local server via `air serve`; `make docker` builds and runs the container locally.

## Coding Style & Naming Conventions
Follow standard Go formatting: tabs for indentation, `gofmt`/`go fmt` layout, and short package names. Keep exported identifiers in `CamelCase`, unexported helpers in `camelCase`, and tests in `*_test.go`. Prefer table-driven tests for API and parser logic. Linting is enforced through `.golangci.yaml`; run `make lint` before opening a PR.

## Testing Guidelines
Unit tests live beside the code they cover. Integration tests are driven by [`main_test.go`](/Users/c/src/iocalebs/patrolbot/main_test.go) and the `.txtar` scripts in [`testdata/integration-tests`](/Users/c/src/iocalebs/patrolbot/testdata/integration-tests). Use `make update` or `make updateall` only when intentionally refreshing golden outputs. Keep new fixtures deterministic by scrubbing timestamps, ports, and other environment-specific values. 

In general, avoid directly editing goldens in integration tests - these should be updated automatically by `make update` or `make updateall`. If any changes to goldens occur as a result of running an update command, review the git diffs to validate them.

## Commit & Pull Request Guidelines
Recent history uses short, imperative commit subjects such as `Add test coverage for patrol command` and `Update trivy version`. Keep commits focused and descriptive. PRs should summarize the behavior change, note config or infrastructure impact, link the relevant issue when applicable, and include screenshots or terminal captures for user-visible CLI or report output changes.

## Security & Configuration Tips
Do not commit plaintext secrets. Use `.env.enc`, `sops`, and the `make decrypt` / `make encrypt` workflow for secret management. Treat `config.yaml` and `templates/*.tmpl` as user-facing contract files; document any breaking changes in the PR.
