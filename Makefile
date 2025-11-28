.PHONY: help build build-prod install test clean

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X 'github.com/anwam/go-template-sh/cmd.Version=$(VERSION)' \
           -X 'github.com/anwam/go-template-sh/cmd.GitCommit=$(GIT_COMMIT)' \
           -X 'github.com/anwam/go-template-sh/cmd.BuildDate=$(BUILD_DATE)'

help:
	@echo "Available targets:"
	@echo "  build      - Build the CLI tool"
	@echo "  build-prod - Build production-ready binary with version info"
	@echo "  install    - Install the CLI tool"
	@echo "  test       - Run tests"
	@echo "  clean      - Clean build artifacts"
	@echo ""
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(GIT_COMMIT)"

build:
	go build -ldflags="$(LDFLAGS)" -o go-template-sh

build-prod:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS) -s -w" -trimpath -o go-template-sh

install:
	go install -ldflags="$(LDFLAGS)"

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

clean:
	rm -f go-template-sh
