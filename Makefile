BIN := bin/cassor
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
TAG := $(shell git describe --tags --exact-match 2>/dev/null)
DIRTY := $(shell test -n "$$(git status --porcelain 2>/dev/null)" && echo .dirty)
VERSION := $(if $(TAG),$(TAG),0.0.0-dev+$(COMMIT)$(DIRTY))
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X github.com/rickcern44/cassor/internal/cmd.Version=$(VERSION) -X github.com/rickcern44/cassor/internal/cmd.Commit=$(COMMIT) -X github.com/rickcern44/cassor/internal/cmd.Date=$(DATE)

.PHONY: build test check site version snapshot release-check

build:
	@mkdir -p bin
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN) .

test:
	go test ./...

check:
	go run . check

site:
	go run . site build

version: build
	$(BIN) version

snapshot:
	goreleaser release --snapshot --clean

release-check:
	goreleaser check
