SHELL := /bin/bash

.DEFAULT_GOAL := build

.PHONY: build run help fmt fmt-check lint test ci tools clean install

BIN_DIR := $(CURDIR)/bin
BIN := $(BIN_DIR)/memelink
CMD := ./cmd/memelink

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT := $(shell git rev-parse --short=12 HEAD 2>/dev/null || echo "")
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
	-X github.com/dedene/memelink-cli/internal/cmd.version=$(VERSION) \
	-X github.com/dedene/memelink-cli/internal/cmd.commit=$(COMMIT) \
	-X github.com/dedene/memelink-cli/internal/cmd.date=$(DATE)

TOOLS_DIR := $(CURDIR)/.tools
GOFUMPT := $(TOOLS_DIR)/gofumpt
GOIMPORTS := $(TOOLS_DIR)/goimports
GOLANGCI_LINT := $(TOOLS_DIR)/golangci-lint

build:
	@mkdir -p $(BIN_DIR)
	@go build -ldflags "$(LDFLAGS)" -o $(BIN) $(CMD)

run: build
	@$(BIN) $(ARGS)

help: build
	@$(BIN) --help

tools:
	@mkdir -p $(TOOLS_DIR)
	@GOBIN=$(TOOLS_DIR) go install mvdan.cc/gofumpt@v0.9.2
	@GOBIN=$(TOOLS_DIR) go install golang.org/x/tools/cmd/goimports@v0.41.0
	@GOBIN=$(TOOLS_DIR) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.8.0

fmt: tools
	@$(GOIMPORTS) -local github.com/dedene/memelink-cli -w .
	@$(GOFUMPT) -w .

fmt-check: tools
	@$(GOIMPORTS) -local github.com/dedene/memelink-cli -w .
	@$(GOFUMPT) -w .
	@git diff --exit-code -- '*.go' go.mod go.sum

lint: tools
	@$(GOLANGCI_LINT) run

test:
	@go test ./...

ci: fmt-check lint test

clean:
	@rm -rf $(BIN_DIR) $(TOOLS_DIR)

install: build
	@cp $(BIN) $(GOPATH)/bin/memelink 2>/dev/null || cp $(BIN) /usr/local/bin/memelink
