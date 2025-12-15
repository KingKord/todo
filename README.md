# Todo gRPC service

Minimal gRPC-based todo service with Postgres storage.

## Requirements
- Go 1.22+ (module uses 1.25)
- Docker + docker-compose (for integration DB)
- protoc, protoc-gen-go, protoc-gen-go-grpc on `PATH`

## Setup
```bash
make generate      # regenerate protobuf stubs
```

## Common commands
- `make lgbt` — generate stubs, format, lint (golangci-lint), run unit + integration tests.
- `make test` — unit tests.
- `make test-integration` — spins up the test Postgres via `docker-compose.test.yml`, runs integration tests, then tears it down.
- `make run` — start the gRPC server locally (needs a running Postgres at `POSTGRES_DSN`).

## Docker
- Local stack (server + db): `docker-compose -f docker-compose.yml up --build`.
- Test DB only (used by `make test-integration`): `docker-compose -f docker-compose.test.yml up -d db`.

## Notes
- Linting uses golangci-lint; the Makefile installs it automatically if missing.
- Postgres default DSN: `postgres://postgres:postgres@localhost:55432/todos?sslmode=disable`.
