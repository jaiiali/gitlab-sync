GOPATH = $(shell go env GOPATH)
GOBIN = $(GOPATH)/bin
PKG_LIST = $(shell go list ./... | grep -v /vendor)

APP_BIN = bin/app

GIT_COMMIT = $(shell git rev-parse --short HEAD)
BUILD_DATE = $(shell date -Iseconds)

LDFLAGS += -X "main.Version=$(GIT_COMMIT)"
LDFLAGS += -X "main.BuildDate=$(BUILD_DATE)"

.PHONY: help
help:
	@ cat $(MAKEFILE_LIST) | grep -e "^[^\.][0-9a-zA-Z\.\_\-]*\:" | awk -F: '{ print $$1 }'

.PHONY: mod
mod:
	go mod tidy

.PHONY: test
test:
	go test -v -race $(PKG_LIST)

.PHONY: coverage
coverage:
	test -e $(GOBIN)/gotestsum || go install gotest.tools/gotestsum@latest
	gotestsum --jsonfile test-output.log --no-summary=skipped --junitfile ./coverage.xml --format short -- -coverprofile=./coverage.txt -covermode=atomic -race $(PKG_LIST)
	go tool cover -func coverage.txt

.PHONY: lint
lint:
	test -e $(GOBIN)/golangci-lint || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint version
	golangci-lint --timeout 5m run

.PHONY: fix
fix:
	test -e $(GOBIN)/golangci-lint || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint version
	golangci-lint --timeout 5m run --fix

.PHONY: build
build:
	CGO_ENABLED=0 go build -tags=containers -ldflags "-s -w $(LDFLAGS)" -o $(APP_BIN) ./...
