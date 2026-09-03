# Repository Guidelines

## Project Structure & Module Organization
- Backend Go service lives in `src/`, with entrypoint `src/main.go` wired through `internal/controllers` (HTTP routes) and `internal/services` (config, DB, provider logic). Shared helpers are under `common/`, constants in `consts/`, and provider adapters in `providers/`.
- Configuration sample is `src/config.example.yaml`; every value has a default, so the server also starts with no config file at all.
- Frontend admin app sits in `ui/` (Vite + Vue 3 + TypeScript). Static assets are under `ui/public/`.
- The binary is self-contained: `migrations/` (per-dialect), `statics/`, and the built UI (`internal/web/dist`, written by `make ui`) are all `go:embed`ed. Do not add runtime reads of files relative to the working directory.
- SQLite is the default database; Postgres stays supported. Schema changes need a migration in **both** `src/migrations/sqlite/` and `src/migrations/postgres/` — `make sql NAME=...` scaffolds both.
- Deployment assets: `Dockerfile` for container builds, `build.sh` for building/pushing and rendering `deploy_k8s.yaml`, and `deploy_k8s.yaml` for Kubernetes.

## Build, Test, and Development Commands
- Backend local run: `cd src && go run .` (SQLite in the working directory; no config needed).
- Backend build: `cd src && make build`; `make all-in-one` builds the UI first so the binary embeds it.
- Backend sanity checks: `cd src && go vet ./...` to catch common issues.
- Backend tests (add `_test.go` first): `cd src && go test ./...`.
- Frontend dev: `cd ui && npm install && npm run dev` for hot-reload.
- Frontend build: `cd ui && npm run build` (type-checks via `vue-tsc` then runs Vite).
- Container build: `./build.sh` or `docker build -t <name> -f Dockerfile .` if you only need the image.

## Coding Style & Naming Conventions
- Go: use `gofmt` (implicit via `go fmt ./...`) and keep packages short, lowercase, and domain-driven (`providers/openai`, `internal/services`). Favor clear, imperative error messages.
- Request/response structs should live near handlers; shared types belong in `common/` or `consts/`.
- Frontend: follow Vue single-file component defaults; keep TypeScript types colocated with usage. Use kebab-case file names for components and camelCase for variables.

## Testing Guidelines
- Place Go tests alongside code as `<name>_test.go`; table-driven tests are preferred for handlers and services.
- Use `go test ./...` before submitting. Aim to cover provider selection logic and request validation paths.
- Frontend currently lacks tests; add Vite/Vitest tests under `ui/src/__tests__` when introducing new behavior.

## Commit & Pull Request Guidelines
- Recent history favors short, imperative commits (e.g., `fix empty content`); keep messages scoped to one logical change and prefix with a verb.
- Include brief PR descriptions: what changed, why, and any user-facing effects. Link issues when applicable and note config changes.
- Add screenshots/gifs for UI-facing updates and sample requests/responses for API changes.
- Do not commit secrets; keep API keys in local `config.yaml` or environment overrides and out of version control.
