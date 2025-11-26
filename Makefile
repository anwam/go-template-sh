.PHONY: help build install test clean

help:
	@echo "Available targets:"
	@echo "  build   - Build the CLI tool"
	@echo "  install - Install the CLI tool"
	@echo "  test    - Run tests"
	@echo "  clean   - Clean build artifacts"

build:
	go build -o go-template-sh

install:
	go install

test:
	go test -v ./...

clean:
	rm -f go-template-sh
