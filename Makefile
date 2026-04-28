GOCACHE := $(CURDIR)/.cache/go-build
GOMODCACHE := $(CURDIR)/.cache/go-mod
GOLANGCI_LINT_CACHE := $(CURDIR)/.cache/golangci-lint
VERSION ?= dev
DIST_DIR := $(CURDIR)/dist
LDFLAGS := -X github.com/pnanadikar/shortcutdeck/internal/appinfo.Version=$(VERSION)

.PHONY: build test lint test-store test-scheduler test-server test-frontend run release

build:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go build -o shortcuts ./cmd/shortcutdeck

test:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go test ./...

lint:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" GOLANGCI_LINT_CACHE="$(GOLANGCI_LINT_CACHE)" golangci-lint run

test-store:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go test ./internal/store/... -v

test-scheduler:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go test ./internal/scheduler/... -v

test-server:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go test ./internal/server/... -v

test-frontend:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go test ./internal/server/... -v

run:
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" go run ./cmd/shortcutdeck

release:
	mkdir -p "$(DIST_DIR)"
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" GOOS=linux GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o "$(DIST_DIR)/shortcutdeck-linux-amd64" ./cmd/shortcutdeck
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" GOOS=darwin GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o "$(DIST_DIR)/shortcutdeck-darwin-amd64" ./cmd/shortcutdeck
	GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)" GOOS=windows GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -o "$(DIST_DIR)/shortcutdeck-windows-amd64.exe" ./cmd/shortcutdeck
