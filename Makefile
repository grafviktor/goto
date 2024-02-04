BUILD_BRANCH  = $(shell git rev-parse --abbrev-ref HEAD)
BUILD_COMMIT  = $(shell git rev-parse --short=8 HEAD)
BUILD_VERSION = v1.0.0
BUILD_DATE    = $(shell date +%Y-%m-%d)
NO_DEBUG_FLAGS = -s -w
# Check if there is no associated tag with this commit, that means that it is a dev build.
BUILD_VERSION_SUFFIX = $(shell git describe --tags --exact-match > /dev/null 2>&1 || echo \\\(dev\\\))
# Use build version and suffix for burning in buildVersion variable:
# For tagged builds - "vX.X.X"
# For non tagged - "vX.X.X (dev)"
BUILD_VERSION_AND_SUFFIX = $(strip $(BUILD_VERSION) $(BUILD_VERSION_SUFFIX))
LD_FLAGS = -ldflags="$(NO_DEBUG_FLAGS) -X main.buildVersion="$(BUILD_VERSION)$(BUILD_VERSION_SUFFIX)" -X main.buildDate=$(BUILD_DATE) -X main.buildCommit=$(BUILD_COMMIT) -X main.buildBranch=$(BUILD_BRANCH)"
DIST_PATH=./dist

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
	go test -coverpkg=./internal/... -race -vet=off -count=1 -coverprofile unit.txt -covermode atomic ./...

# unit-test-report: display unit coverage report in html format. This option is hidden from make help menu.
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

## build: create binary in ./build/dist folder for your current platform. Use this option if you build it for personal use.
.PHONY: build
build:
	@-rm -r $(DIST_PATH)/gg
	@echo 'Building'
	go build $(LD_FLAGS) -o $(DIST_PATH)/gg ./cmd/goto/*.go

## package: create rpm package and place it into ./build/dist folder.
.PHONY: package
package:
	@-rm -r $(DIST_PATH)/*.rpm $(DIST_PATH)/*.deb
	@echo 'Build rpm package'
# Use cut to convert version from 'vX.X.X' to 'X.X.X'
	@DOCKER_BUILDKIT=1 docker build --build-arg VERSION=$(shell echo $(BUILD_VERSION) | cut -c 2-) -f build/rpm/Dockerfile --output ./dist .
	@echo 'Build deb package'
	@DOCKER_BUILDKIT=1 docker build --build-arg VERSION=$(shell echo $(BUILD_VERSION) | cut -c 2-) -f build/deb/Dockerfile --output ./dist .

## dist: create binaries for all supported platforms in ./build/dist folder. Archive all binaries with zip.
.PHONY: dist
dist:
	@-rm -r $(DIST_PATH)/gg-*
	@-rm -r $(DIST_PATH)/*.zip
	@echo 'Creating binary files'
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build $(LD_FLAGS) -o $(DIST_PATH)/gg-mac     ./cmd/goto/*.go
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build $(LD_FLAGS) -o $(DIST_PATH)/gg-lin     ./cmd/goto/*.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LD_FLAGS) -o $(DIST_PATH)/gg-win.exe ./cmd/goto/*.go
	@mkdir $(DIST_PATH)/goto-$(BUILD_VERSION)/
	@cp $(DIST_PATH)/gg-mac $(DIST_PATH)/gg-lin $(DIST_PATH)/gg-win.exe $(DIST_PATH)/goto-$(BUILD_VERSION)
	@cd $(DIST_PATH) && zip -r goto-$(BUILD_VERSION).zip goto-$(BUILD_VERSION)
	@rm -r $(DIST_PATH)/goto-$(BUILD_VERSION)