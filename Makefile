TICP_PKG := github.com/pingcap/ticp

TEST_PKGS := $(shell find . -iname "*_test.go" -exec dirname {} \; | \
                     sort -u | sed -e "s/^\./github.com\/pingcap\/ticp/")
INTEGRATION_TEST_PKGS := $(shell find . -iname "*_test.go" -exec dirname {} \; | \
                     sort -u | sed -e "s/^\./github.com\/pingcap\/ticp/" | grep -E "tests")
BASIC_TEST_PKGS := $(filter-out $(INTEGRATION_TEST_PKGS),$(TEST_PKGS))

PACKAGES := go list ./...
PACKAGE_DIRECTORIES := $(PACKAGES) | sed 's|$(TICP_PKG)/||'
GOCHECKER := awk '{ print } END { if (NR > 0) { exit 1 } }'
OVERALLS := overalls

BUILD_BIN_PATH := $(shell pwd)/bin
GO_TOOLS_BIN_PATH := $(shell pwd)/.tools/bin
PATH := $(GO_TOOLS_BIN_PATH):$(PATH)
SHELL := env PATH='$(PATH)' GOBIN='$(GO_TOOLS_BIN_PATH)' $(shell which bash)

FAILPOINT_ENABLE  := $$(find $$PWD/ -type d | grep -vE "\.git" | xargs failpoint-ctl enable)
FAILPOINT_DISABLE := $$(find $$PWD/ -type d | grep -vE "\.git" | xargs failpoint-ctl disable)

DEADLOCK_ENABLE := $$(\
						find . -name "*.go" \
						| xargs -n 1 sed -i.bak 's/sync\.RWMutex/deadlock.RWMutex/;s/sync\.Mutex/deadlock.Mutex/' && \
						find . -name "*.go" | xargs grep -lE "(deadlock\.RWMutex|deadlock\.Mutex)" \
						| xargs goimports -w)
DEADLOCK_DISABLE := $$(\
						find . -name "*.go" \
						| xargs -n 1 sed -i.bak 's/deadlock\.RWMutex/sync.RWMutex/;s/deadlock\.Mutex/sync.Mutex/' && \
						find . -name "*.go" | xargs grep -lE "(sync\.RWMutex|sync\.Mutex)" \
						| xargs goimports -w && \
						find . -name "*.bak" | xargs rm && \
						go mod tidy)

BUILD_FLAGS ?=
BUILD_TAGS ?=
BUILD_CGO_ENABLED := 0
TICP_EDITION ?= Enterprise

# Ensure TICP_EDITION is set to Community or Enterprise before running build process.
ifneq "$(TICP_EDITION)" "Community"
ifneq "$(TICP_EDITION)" "Enterprise"
  $(error Please set the correct environment variable TICP_EDITION before running `make`)
endif
endif

ifneq ($(SWAGGER), 0)
	BUILD_TAGS += swagger_server
endif

ifeq ("$(WITH_RACE)", "1")
	BUILD_FLAGS += -race
	BUILD_CGO_ENABLED := 1
endif

#LDFLAGS += -X "$(TICP_PKG)/server/versioninfo.PDReleaseVersion=$(shell git describe --tags --dirty --always)"
#LDFLAGS += -X "$(TICP_PKG)/server/versioninfo.PDBuildTS=$(shell date -u '+%Y-%m-%d %I:%M:%S')"
#LDFLAGS += -X "$(TICP_PKG)/server/versioninfo.PDGitHash=$(shell git rev-parse HEAD)"
#LDFLAGS += -X "$(TICP_PKG)/server/versioninfo.PDGitBranch=$(shell git rev-parse --abbrev-ref HEAD)"
#LDFLAGS += -X "$(TICP_PKG)/server/versioninfo.PDEdition=$(TICP_EDITION)"

GOVER_MAJOR := $(shell go version | sed -E -e "s/.*go([0-9]+)[.]([0-9]+).*/\1/")
GOVER_MINOR := $(shell go version | sed -E -e "s/.*go([0-9]+)[.]([0-9]+).*/\2/")
GO111 := $(shell [ $(GOVER_MAJOR) -gt 1 ] || [ $(GOVER_MAJOR) -eq 1 ] && [ $(GOVER_MINOR) -ge 11 ]; echo $$?)
ifeq ($(GO111), 1)
  $(error "go below 1.11 does not support modules")
endif

default: build

all: dev

dev: build check tools test

ci: build check basic-test

build: ticp-server

tools:

TICP_SERVER_DEP :=
ifneq ($(SWAGGER), 0)
	TICP_SERVER_DEP += swagger-spec
endif

ticp-server: export GO111MODULE=on
ticp-server: ${TICP_SERVER_DEP}
	CGO_ENABLED=$(BUILD_CGO_ENABLED) go build $(BUILD_FLAGS) -gcflags '$(GCFLAGS)' -ldflags '$(LDFLAGS)' -tags "$(BUILD_TAGS)" -o $(BUILD_BIN_PATH)/ticp-server micro/ticp-server/main.go

ticp-server-basic: export GO111MODULE=on
ticp-server-basic:
	SWAGGER=0 $(MAKE) ticp-server

# dependent
install-go-tools: export GO111MODULE=on
install-go-tools:
	@mkdir -p $(GO_TOOLS_BIN_PATH)
	@which golangci-lint >/dev/null 2>&1 || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GO_TOOLS_BIN_PATH) v1.38.0
	@grep '_' tools.go | sed 's/"//g' | awk '{print $$2}' | xargs go install

swagger-spec: export GO111MODULE=on
swagger-spec: install-go-tools
	go mod vendor
	#swag init --parseVendor -generalInfo server/api/router.go --exclude vendor/github.com/pingcap/tidb-dashboard --output docs/swagger
	#go mod tidy
	rm -rf vendor

# Tools

test: install-go-tools
	# testing...
	@$(DEADLOCK_ENABLE)
	@$(FAILPOINT_ENABLE)
	CGO_ENABLED=1 GO111MODULE=on go test -race -cover $(TEST_PKGS) || { $(FAILPOINT_DISABLE); $(DEADLOCK_DISABLE); exit 1; }
	@$(FAILPOINT_DISABLE)
	@$(DEADLOCK_DISABLE)

basic-test:
	@$(FAILPOINT_ENABLE)
	GO111MODULE=on go test $(BASIC_TEST_PKGS) || { $(FAILPOINT_DISABLE); exit 1; }
	@$(FAILPOINT_DISABLE)

test-with-cover: install-go-tools
	# testing...
	@$(FAILPOINT_ENABLE)
	for PKG in $(TEST_PKGS); do\
		set -euo pipefail;\
		CGO_ENABLED=1 GO111MODULE=on go test -race -covermode=atomic -coverprofile=coverage.tmp -coverpkg=./... $$PKG  2>&1 | grep -v "no packages being tested" && tail -n +2 coverage.tmp >> covprofile || { $(FAILPOINT_DISABLE); rm coverage.tmp && exit 1;}; \
		rm coverage.tmp;\
	done
	@$(FAILPOINT_DISABLE)

check: install-go-tools check-all check-plugin errdoc check-testing-t docker-build-test

check-all: static lint tidy
	@echo "checking"

check-plugin:
	@echo "checking plugin"

static: export GO111MODULE=on
static: install-go-tools
	@ # Not running vet and fmt through metalinter becauase it ends up looking at vendor
	gofmt -s -l -d $$($(PACKAGE_DIRECTORIES)) 2>&1 | $(GOCHECKER)
	golangci-lint run $$($(PACKAGE_DIRECTORIES))

lint: install-go-tools
	@echo "linting"
	revive -formatter friendly -config revive.toml $$($(PACKAGES))

tidy:
	@echo "go mod tidy"
	GO111MODULE=on go mod tidy
	git diff --quiet go.mod go.sum

errdoc: install-go-tools
	@echo "generator errors.toml"
	#./scripts/check-errdoc.sh

docker-build-test:
	$(eval DOCKER_PS_EXIT_CODE=$(shell docker ps > /dev/null 2>&1 ; echo $$?))
	@if [ $(DOCKER_PS_EXIT_CODE) -ne 0 ]; then \
	echo "Encountered problem while invoking docker cli. Is the docker daemon running?"; \
	fi
	docker build --no-cache -t pingcap/ticp .

check-testing-t:
	#./scripts/check-testing-t.sh

clean-test:
	# Cleaning test tmp...
	rm -rf /tmp/test_pd*
	rm -rf /tmp/pd-tests*
	rm -rf /tmp/test_etcd*

clean-build:
	# Cleaning building files...
	rm -rf $(BUILD_BIN_PATH)
	rm -rf $(GO_TOOLS_BIN_PATH)

deadlock-enable: install-go-tools
	# Enabling deadlock...
	@$(DEADLOCK_ENABLE)

deadlock-disable: install-go-tools
	# Disabling deadlock...
	@$(DEADLOCK_DISABLE)

failpoint-enable: install-go-tools
	# Converting failpoints...
	@$(FAILPOINT_ENABLE)

failpoint-disable: install-go-tools
	# Restoring failpoints...
	@$(FAILPOINT_DISABLE)

clean: failpoint-disable deadlock-disable clean-test clean-build

.PHONY: all ci vendor tidy clean-test clean-build clean
