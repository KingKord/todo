APP_NAME=todo
GRPC_ADDR?=:50051
POSTGRES_DSN?=postgres://postgres:postgres@localhost:55432/todos?sslmode=disable
COMPOSE_TEST_FILE?=docker-compose.test.yml
GOLANGCI_LINT_VERSION?=v1.62.2
GOLANGCI_LINT_BIN?=$(HOME)/go/bin/golangci-lint

.PHONY: lgbt
lgbt: generate fmt lint test test-integration

.PHONY: generate
generate: proto

.PHONY: proto
proto:
	protoc -I api \
		--go_out=internal/gen --go_opt=paths=source_relative \
		--go-grpc_out=internal/gen --go-grpc_opt=paths=source_relative \
		api/todo/v1/todo.proto

.PHONY: run
run:
	GRPC_ADDR=$(GRPC_ADDR) POSTGRES_DSN=$(POSTGRES_DSN) go run ./cmd/server

.PHONY: build
build:
	go build -o bin/$(APP_NAME) ./cmd/server

.PHONY: test
test:
	go test ./...

.PHONY: golangci-lint
golangci-lint:
	@if [ ! -x "$(GOLANGCI_LINT_BIN)" ]; then \
		echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."; \
		GOBIN=$$(dirname "$(GOLANGCI_LINT_BIN)") go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	fi

.PHONY: lint
lint: golangci-lint
	$(GOLANGCI_LINT_BIN) run ./...

.PHONY: fmt
fmt:
	gofmt -w ./cmd ./internal ./integration

.PHONY: test-integration
test-integration: compose-down compose-up
	POSTGRES_DSN=$(POSTGRES_DSN) go test -tags=integration ./integration
	make compose-down

.PHONY: compose-up
compose-up:
	docker-compose -f $(COMPOSE_TEST_FILE) up -d db

.PHONY: compose-down
compose-down:
	docker-compose -f $(COMPOSE_TEST_FILE) down -v
