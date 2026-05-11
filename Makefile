GO ?= go
BUF ?= buf
COMPOSE ?= docker compose

.PHONY: help
help:
	@echo "make proto       - generate gRPC code from proto/"
	@echo "make tidy        - go mod tidy"
	@echo "make lint        - run golangci-lint"
	@echo "make test        - run unit tests with race detector"
	@echo "make build       - build all service binaries into bin/"
	@echo "make up          - start full stack via docker compose"
	@echo "make down        - stop and remove containers"
	@echo "make logs        - tail container logs"
	@echo "make ps          - list running containers"
	@echo "make clean       - remove generated code and binaries"

.PHONY: proto
proto:
	$(BUF) generate

.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	$(GO) test -race -count=1 ./...

.PHONY: build
build:
	mkdir -p bin
	$(GO) build -o bin/auth-service ./services/auth-service/cmd/server
	$(GO) build -o bin/user-service ./services/user-service/cmd/server

.PHONY: up
up:
	$(COMPOSE) up -d --build

.PHONY: down
down:
	$(COMPOSE) down -v

.PHONY: logs
logs:
	$(COMPOSE) logs -f --tail=200

.PHONY: ps
ps:
	$(COMPOSE) ps

.PHONY: clean
clean:
	rm -rf bin/ gen/
