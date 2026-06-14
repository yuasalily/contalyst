BINARY  := contalyst
PKG     := ./...
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: build run test test-unit test-functional test-e2e cover vet fmt fmt-check tidy lint snapshot clean install help

build: ## Build the binary with version info
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

run: ## Build and run
	go run .

test: ## Run unit + functional tests (no daemon required)
	go test -race $(PKG)

test-unit: ## Run only unit tests (pure functions)
	go test -race -run TestUnit $(PKG)

test-functional: ## Run only functional tests (model-driven, no daemon)
	go test -race -run TestFunctional $(PKG)

test-e2e: ## Run end-to-end tests against the live Docker daemon
	go test -tags e2e -run TestE2E -v -timeout 300s $(PKG)

cover: ## Unit + functional coverage report
	go test -covermode=atomic -coverprofile=coverage.out $(PKG)
	go tool cover -func=coverage.out | tail -1

vet: ## Static analysis (incl. e2e build tag)
	go vet $(PKG)
	go vet -tags e2e $(PKG)

fmt: ## Format code
	gofmt -w .

fmt-check: ## Fail if any file is not gofmt-formatted
	@test -z "$$(gofmt -l .)" || (echo "unformatted:"; gofmt -l .; exit 1)

lint: fmt-check vet ## Run the lint gate (gofmt + go vet)

tidy: ## Sync go.mod/go.sum
	go mod tidy

snapshot: ## Build a local snapshot release with GoReleaser (no publish)
	goreleaser release --snapshot --clean --skip=publish

install: ## Install to GOBIN
	go install -ldflags "$(LDFLAGS)" .

clean:
	rm -f $(BINARY) coverage.out
	rm -rf dist
	go clean

help: ## List targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'
