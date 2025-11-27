.PHONY: help build build-prod install test clean

help:
	@echo "Available targets:"
	@echo "  build      - Build the CLI tool"
	@echo "  build-prod - Build production-ready binary"
	@echo "  install    - Install the CLI tool"
	@echo "  test    - Run tests"
	@echo "  clean   - Clean build artifacts"

build:
	go build -o go-template-sh

build-prod:
	CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o go-template-sh

install:
	go install

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

clean:
	rm -f go-template-sh
