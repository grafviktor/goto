BUILD_BRANCH  = $(shell git rev-parse --abbrev-ref HEAD)
BUILD_COMMIT  = $(shell git rev-parse --short=8 HEAD)
BUILD_VERSION = v1.5.0
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
	@echo 'Run linter'
	@golangci-lint run

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit:
	@echo 'If something is not right, make sure these utilities are installed:'
	@echo '  * If gofumpt is not installed, then: go install mvdan.cc/gofumpt@latest'
	@echo '  * For staticcheck: go install honnef.co/go/tools/cmd/staticcheck@2025.1.1'
	@echo '  * If goimports is not installed, then: go install golang.org/x/tools/cmd/goimports@latest'
	@echo '  * For golangci-lint: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin v2.2.1'
	@echo 'Tidy and verify module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Format code...'
	gofumpt -l -w -extra ./
	goimports -local "github.com/grafviktor/goto" -w .
	@echo 'Vet code...'
	go vet ./...
	staticcheck ./...
	@echo 'Lint code...'
	golangci-lint run
	@$(MAKE) test

## test: run unit tests
.PHONY: test
test:
	@echo 'Run unit tests'
	go test -coverpkg=./internal/... -race -vet=off -count=1 -coverprofile unit.txt -covermode atomic ./...

# unit-test-report: display unit coverage report in html format. This option is hidden from make help menu.
.PHONY: unit-test-report
unit-test-report:
	@echo 'The report will be opened in the browser'
	go tool cover -html unit.txt

## run: delete logs and run debug
.PHONY: run
run:
	@echo 'Run debug build'
	@echo 'To pass app arguments use: make run ARGS="-h"'
	go run cmd/goto/* $(ARGS)

## build: create binary in ./dist folder for your current platform. Use this option if you build it for personal use.
.PHONY: build
build:
	@-rm -f $(DIST_PATH)/gg 2>/dev/null
	@echo 'Build'
	go build $(LD_FLAGS) -o $(DIST_PATH)/gg ./cmd/goto/*.go

## package: create rpm and deb packages and place them into ./dist folder.
.PHONY: package
package:
	@-rm -r $(DIST_PATH)/*.rpm $(DIST_PATH)/*.deb 2>/dev/null
	@echo 'Build rpm package'
# Use cut to convert version from 'vX.X.X' to 'X.X.X'
# Nice trick to avoid docker cache problem. Here we're passing an extra variable which contains last commit
# if there is a new commit which is different which saved in cache, then the build will be invalidated
# docker build --build-arg CACHEBUST=`git rev-parse ${GITHUB_REF}` ... or use --no-cache
	@DOCKER_BUILDKIT=1 BUILDKIT_PROGRESS=plain docker build --build-arg VERSION=$(shell echo $(BUILD_VERSION) | cut -c 2-) --build-arg BRANCH=$(BUILD_BRANCH) -f build/rpm/Dockerfile --output ./dist .
	@echo 'Build deb package'
	@DOCKER_BUILDKIT=1 BUILDKIT_PROGRESS=plain docker build --build-arg VERSION=$(shell echo $(BUILD_VERSION) | cut -c 2-) --build-arg BRANCH=$(BUILD_BRANCH) -f build/deb/Dockerfile --output ./dist .

## dist: create binaries for all supported platforms in ./dist folder. Archive all binaries with zip.
.PHONY: dist
dist:
	@echo 'Generate icon resource for windows binary.'
	@echo 'Do not forget to: "go install github.com/akavel/rsrc@latest"'
	@cd build/exe && rsrc -arch amd64 -ico icon.ico -o icon_windows_amd64.syso && mv icon_windows_amd64.syso ../../cmd/goto
	@-rm -f $(DIST_PATH)/gg-* 2>/dev/null
	@-rm -f $(DIST_PATH)/*.zip 2>/dev/null
	@echo 'Create binary files'
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build $(LD_FLAGS) -o $(DIST_PATH)/gg-mac-x86 ./cmd/goto/
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build $(LD_FLAGS) -o $(DIST_PATH)/gg-mac-arm ./cmd/goto/
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build $(LD_FLAGS) -o $(DIST_PATH)/gg-lin     ./cmd/goto/
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LD_FLAGS) -o $(DIST_PATH)/gg-win.exe ./cmd/goto/
	@rm cmd/goto/icon_windows_amd64.syso
	@mkdir $(DIST_PATH)/goto-$(BUILD_VERSION)/
	@cp $(DIST_PATH)/gg-mac-arm $(DIST_PATH)/gg-mac-x86 $(DIST_PATH)/gg-lin $(DIST_PATH)/gg-win.exe $(DIST_PATH)/goto-$(BUILD_VERSION)
	@cd $(DIST_PATH) && zip -r goto-$(BUILD_VERSION).zip goto-$(BUILD_VERSION)
	@rm -rf $(DIST_PATH)/goto-$(BUILD_VERSION)

## clean: remove ./dist folder with all its contents.
.PHONY: clean
clean:
	@echo 'Clean'
	@-rm -rf $(DIST_PATH) 2>/dev/null
	@echo 'Done'