APP_NAME := upass
GO_CMD := go
LDFLAGS := -ldflags="-s -w"

.PHONY: all build test lint clean release help

all: help

build:
	@echo "🔨 Building $(APP_NAME)..."
	$(GO_CMD) build $(LDFLAGS) -o $(APP_NAME) .

test:
	@echo "🧪 Running tests..."
	$(GO_CMD) test -v -race ./...

lint:
	@echo "🔍 Running linter (golangci-lint)..."
	golangci-lint run ./...

clean:
	@echo "🧹 Cleaning up..."
	rm -f $(APP_NAME)
	rm -f $(APP_NAME).exe

release:
	@echo "📦 Building release binaries..."
	GOOS=linux GOARCH=amd64 $(GO_CMD) build $(LDFLAGS) -o $(APP_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 $(GO_CMD) build $(LDFLAGS) -o $(APP_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GO_CMD) build $(LDFLAGS) -o $(APP_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 $(GO_CMD) build $(LDFLAGS) -o $(APP_NAME)-windows-amd64.exe .
	@echo "✅ Binaries built successfully!"

help:
	@echo "Available commands:"
	@echo "  make build   - Compile the application"
	@echo "  make test    - Run all tests with race detector"
	@echo "  make lint    - Run golangci-lint"
	@echo "  make clean   - Remove compiled binaries"
	@echo "  make release - Build binaries for Linux, macOS, and Windows"
