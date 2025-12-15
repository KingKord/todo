# Repository Guidelines

## Project Structure & Module Organization
- gRPC Todo entrypoint: `cmd/server/main.go`; server wiring and health checks: `internal/server/`.
- API definitions live in `api/todo/v1/todo.proto`; generated stubs are in `internal/gen/todo/v1/`.
- Database access sits in `internal/todo/`; migrations are embedded from `internal/storage/migrations/`.
- Integration tests (tagged `integration`) live in `integration/`; Docker orchestration is in `docker-compose.yml`.
- Use `configs/` and `scripts/` for env samples and helper scripts when you add them.

## Build, Test, and Development Commands
- Run locally against Postgres (default DSN for compose): `make run` (uses `GRPC_ADDR` and `POSTGRES_DSN` envs).
- Build binary: `make build` → `bin/server`.
- Unit tests: `make test`; integration (needs Docker): `make test-integration` (starts compose `db`).
- Format: `make fmt`; add `go vet ./...` before pushing.

## Coding Style & Naming Conventions
- Rely on `gofmt` defaults (tabs, grouped imports); avoid manual styling drift.
- Packages are lower-case, no underscores; exported types/functions use PascalCase, locals use mixedCaps.
- Match filenames to their main responsibility (`todo_service.go`, `repository.go`, `todo.pb.go`); keep services thin and push DB logic into repositories.

## Testing Guidelines
- Use Go’s built-in `testing` package; name files `*_test.go` and entries `TestXxx`.
- Prefer table-driven tests; keep fixtures in `testdata/` or embedded, and avoid external network calls outside Dockerized dependencies.
- Run `go test ./... -cover` for unit coverage; integration requires Docker and the `integration` build tag.
- Add regression cases with bug fixes; consider benchmarks when performance-sensitive paths change.

## Commit & Pull Request Guidelines
- Use clear, action-oriented commit subjects (imperative); bodies explain motivation and scope. Conventional Commits are welcome.
- PRs include a summary, rationale, and verification steps (`make test`, `make test-integration`); attach logs or screenshots if relevant.
- Link tracking issues where available; keep PRs small and focused on one change-set.
- Do not commit editor cruft or secrets; add deps via `go get`, commit `go.mod`/`go.sum`, and regenerate stubs from `proto/` when the schema changes.

## Setup & Tooling Tips
- Develop with Go ≥1.22 to match the Docker base; module name is `todo`.
- Run `go mod tidy` before container builds to refresh `go.sum`; Docker requires dependencies fetched ahead of time.
- Default compose DSN: `postgres://postgres:postgres@localhost:55432/todos?sslmode=disable`.
