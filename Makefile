.PHONY: build test clean install doctor

VERSION ?= 0.1.0
BINARY  := bin/archon

## build: Compile archon binary
build:
	go build -ldflags "-X github.com/sirrryasir/archon/cmd.Version=$(VERSION)" -o $(BINARY) .

## test: Run all archon tests (excludes third-party opensource/ directory)
test:
	go test ./cmd/... ./mcp/... ./modes/... ./ai/... ./config/... ./session/... ./tui/... -v

## test-short: Run tests without verbose output
test-short:
	go test ./cmd/... ./mcp/... ./modes/... ./ai/... ./config/... ./session/... ./tui/...

## clean: Remove built binary
clean:
	rm -f $(BINARY)

## install: Install archon to GOPATH/bin
install:
	go install -ldflags "-X github.com/sirrryasir/archon/cmd.Version=$(VERSION)" .

## doctor: Run archon doctor after building
doctor: build
	$(BINARY) doctor

## e2e: Run end-to-end integration tests
e2e:
	./scripts/e2e_test.sh

## help: Show available targets
help:
	@grep -E '^## ' Makefile | sed 's/## /  /'
