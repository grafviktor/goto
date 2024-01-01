BUILD_BRANCH  = $(shell git rev-parse --abbrev-ref HEAD)
BUILD_COMMIT  = $(shell git rev-parse --short=8 HEAD)
BUILD_VERSION = v0.4.0
BUILD_DATE    = $(shell date +%Y-%m-%d)
NO_DEBUG_FLAGS = -s -w
# Check if there is no associated tag with this commit, that means that it is a dev build.
BUILD_VERSION_SUFFIX = $(shell git describe --tags --exact-match > /dev/null 2>&1 || echo \\\(dev\\\))
# Use build version and suffix for burning in buildVersion variable:
# For tagged builds - "vX.X.X"
# For non tagged - "vX.X.X (dev)"
BUILD_VERSION_AND_SUFFIX = $(strip $(BUILD_VERSION) $(BUILD_VERSION_SUFFIX))
LD_FLAGS = -ldflags="$(NO_DEBUG_FLAGS) -X main.buildVersion="$(BUILD_VERSION)$(BUILD_VERSION_SUFFIX)" -X main.buildDate=$(BUILD_DATE) -X main.buildCommit=$(BUILD_COMMIT) -X main.buildBranch=$(BUILD_BRANCH)"

## help: print this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## lint: run linter
.PHONY: lint
lint:
	@echo 'Running linter'
	@golangci-lint run

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Formatting code...'
	gofumpt -l -w -extra ./
	goimports -w -local github.com/grafviktor/goto .
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Linting code...'
	golangci-lint run
	@$(MAKE) test

## test: run unit tests
.PHONY: test
test:
	@echo 'Running unit tests'
# By using -coverpkg=$$(go list ./internal/..|grep -v mock) we exclude "mock" folder from codecov report
	go test -coverpkg=$$(go list ./internal/..|grep -v mock) -race -vet=off -count=1 -coverprofile unit.txt -covermode atomic ./...

## unit-test-report: display unit coverage report in html format
.PHONY: unit-test-report
unit-test-report:
	@echo 'The report will be opened in the browser'
	go tool cover -html unit.txt

## run: delete logs and run debug
.PHONY: run
run:
	@echo 'Running debug build'
	@-rm debug.log 2>/dev/null
	@echo 'To pass app arguments use: make run ARGS="-h"'
	go run cmd/goto/* $(ARGS)

## build: create binaries for all supported platforms in ./build folder. Archive all binaries with zip.
.PHONY: build
build:
	@-rm -r ./build/*
	@echo 'Creating binary files'
	GOOS=darwin  GOARCH=amd64 go build $(LD_FLAGS) -o ./build/gg-mac     ./cmd/goto/*.go
	GOOS=linux   GOARCH=amd64 go build $(LD_FLAGS) -o ./build/gg-lin     ./cmd/goto/*.go
	GOOS=windows GOARCH=amd64 go build $(LD_FLAGS) -o ./build/gg-win.exe ./cmd/goto/*.go
	@mkdir ./build/goto-$(BUILD_VERSION)/
	@cp ./build/gg* ./build/goto-$(BUILD_VERSION)
	@cd ./build && zip -r goto-$(BUILD_VERSION).zip goto-$(BUILD_VERSION)
	@rm -r ./build/goto-$(BUILD_VERSION)

## build-quick: create binary in ./build folder for your current platform
.PHONY: build-quick
build-quick:
	go build $(LD_FLAGS) -o ./build/gg ./cmd/goto/*.go
	@echo 'Creating build'