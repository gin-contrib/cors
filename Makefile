GO ?= go
TOOLS_MOD := -modfile=$(CURDIR)/go.tools.mod
LINT_CONFIG := $(CURDIR)/.golangci.yml
GO_PACKAGES := ./... ./_examples/...
EXAMPLE_MODULES :=

## lint: run golangci-lint to check for issues
lint:
	$(GO) tool $(TOOLS_MOD) golangci-lint run --config=$(LINT_CONFIG) $(GO_PACKAGES)

## fmt: format go files using golangci-lint
fmt:
	$(GO) tool $(TOOLS_MOD) golangci-lint fmt

## test: run tests
test:
	$(GO) test -v -covermode=atomic -coverprofile=coverage.txt $(GO_PACKAGES) && echo "\n==>\033[32m Ok\033[m\n" || exit 1

## coverage: view test coverage in browser
coverage: test
	$(GO) tool cover -html=coverage.txt

## mod-download: download go module dependencies
mod-download:
	$(GO) mod download

## mod-tidy: tidy go module dependencies
mod-tidy:
	$(GO) mod tidy

## mod-verify: verify go module dependencies
mod-verify:
	$(GO) mod verify

## install-tools: download tool dependencies
install-tools:
	$(GO) mod download $(TOOLS_MOD)

## help: print this help message
help:
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: lint fmt test coverage mod-download mod-tidy mod-verify install-tools help
