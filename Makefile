# EnvSync Makefile
# Usage: make <target>

BINARY   := envsync
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  := -ldflags="-s -w -X main.version=$(VERSION)"
GOFLAGS  := -race

.PHONY: all build clean test lint install run-demo help

## all: Build and test
all: build test

## build: Compile the binary
build:
	@echo "→ Building $(BINARY) $(VERSION)..."
	go build $(LDFLAGS) -o $(BINARY) ./...
	@echo "✔ Binary: ./$(BINARY)"

## build-all: Cross-compile for Linux, macOS, Windows
build-all:
	@echo "→ Cross-compiling..."
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64 ./...
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64 ./...
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 ./...
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-amd64.exe ./...
	@echo "✔ Binaries in ./dist/"

## install: Install binary to $GOPATH/bin
install:
	go install $(LDFLAGS) ./...
	@echo "✔ Installed to $(shell go env GOPATH)/bin/$(BINARY)"

## test: Run all unit tests
test:
	@echo "→ Running tests..."
	go test $(GOFLAGS) ./... -v -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -1

## test-cover: Open coverage HTML report
test-cover: test
	go tool cover -html=coverage.out

## lint: Run golangci-lint
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

## clean: Remove build artifacts
clean:
	rm -f $(BINARY) coverage.out
	rm -rf dist/ .envsync/

## demo-init: Set up demo environment files
demo-init:
	@echo "→ Setting up demo files..."
	cp .env.example .env.dev    2>/dev/null || true
	cp .env.example .env.staging 2>/dev/null || true
	cp .env.example .env.production 2>/dev/null || true
	@echo "✔ Created .env.dev, .env.staging, .env.production"
	@echo "  Edit them with different values to simulate drift."

## demo-key: Generate an encryption key
demo-key:
	@echo "→ Generating ENVSYNC_KEY..."
	@echo "export ENVSYNC_KEY=$$(openssl rand -base64 32)"
	@echo ""
	@echo "  Copy the line above and run it in your shell."

## run-diff: Quick demo diff
run-diff: build
	./$(BINARY) diff dev staging

## run-audit: Quick demo audit
run-audit: build
	./$(BINARY) audit --env staging

## run-validate: Quick demo validate
run-validate: build
	./$(BINARY) validate --env dev

## help: Show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'
