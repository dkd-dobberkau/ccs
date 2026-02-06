.PHONY: build install clean test build-all lint fmt deps help

BINARY_NAME=ccs
VERSION?=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
BUILD_DIR=build
INSTALL_DIR?=$(HOME)/.local/bin

LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod

build:
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/ccs/

install: build
	@mkdir -p $(INSTALL_DIR)
	cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed to $(INSTALL_DIR)/$(BINARY_NAME)"

test:
	$(GOTEST) -v ./...

clean:
	rm -rf $(BUILD_DIR)

deps:
	$(GOMOD) tidy
	$(GOMOD) download

build-all: clean
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/ccs/
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/ccs/
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/ccs/
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/ccs/
	@echo "Binaries built in $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

lint:
	golangci-lint run ./...

fmt:
	$(GOCMD) fmt ./...

help:
	@echo "Available targets:"
	@echo "  build     - Build for current platform"
	@echo "  install   - Build and install to ~/.local/bin"
	@echo "  test      - Run tests"
	@echo "  clean     - Remove build artifacts"
	@echo "  deps      - Update dependencies"
	@echo "  build-all - Build for all platforms"
	@echo "  lint      - Run linter"
	@echo "  fmt       - Format code"
