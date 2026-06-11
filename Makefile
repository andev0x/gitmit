BINARY_NAME := gitmit

ifeq ($(OS),Windows_NT)
EXE_NAME := $(BINARY_NAME).exe
INSTALL_DIR := $(USERPROFILE)\bin
else
EXE_NAME := $(BINARY_NAME)
INSTALL_DIR := /usr/local/bin
endif

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

BIN_DIR := bin
BIN_PATH := $(BIN_DIR)/$(EXE_NAME)

.PHONY: all
all: test build

.PHONY: prepare
prepare:
	mkdir -p $(BIN_DIR)

.PHONY: build
build: prepare
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BIN_PATH) .

.PHONY: build-all
build-all:
	@echo "Building for multiple platforms..."
	mkdir -p dist

	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .

.PHONY: test
test:
	go test -v ./...

.PHONY: test-coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: fmt
fmt:
	go fmt -s ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: deps
deps:
	go mod download
	go mod tidy

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
	rm -rf dist
	rm -f coverage.out coverage.html

.PHONY: run
run: build
	./$(BIN_PATH)

.PHONY: install
install:
	go install .

.PHONY: dev-setup
dev-setup:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go mod download

.PHONY: help
help:
	@echo "Targets:"
	@echo "  build"
	@echo "  build-all"
	@echo "  test"
	@echo "  test-coverage"
	@echo "  fmt"
	@echo "  vet"
	@echo "  lint"
	@echo "  deps"
	@echo "  clean"
	@echo "  run"
	@echo "  install"
	@echo "  dev-setup"
