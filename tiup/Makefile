PROJECT=tiem/tiup


.PHONY: cluster

default: all

GOPATH ?= $(shell go env GOPATH)
# Ensure GOPATH is set before running build process.
ifeq "$(GOPATH)" ""
ifneq ($(MAKECMDGOALS),install)
	$(error Please set the environment variable GOPATH before running `make`)
else
	GOPATH := toinstall
endif

endif

CURDIR := $(shell pwd)
path_to_add := $(addsuffix /bin,$(subst :,/bin:,$(GOPATH))):$(PWD)/tools/bin
export PATH := $(path_to_add):$(PATH)

GOOS    := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
GOARCH  := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
GOENV   := GO111MODULE=on CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH)
GO      := $(GOENV) go
GOBUILD := $(GO) build $(BUILD_FLAGS)
GOTEST    := CGO_ENABLED=0 $(GO) test -p 4
OVERALLS  := CGO_ENABLED=1 GO111MODULE=on overalls

PACKAGE_LIST  := go list ./...| grep -vE "cmd" | grep -vE "test"
PACKAGES  := $$($(PACKAGE_LIST))
PACKAGE_DIRECTORIES := $(PACKAGE_LIST) | sed 's|github.com/pingcap-inc/$(PROJECT)/||'
FILES     := $$(find $$($(PACKAGE_DIRECTORIES)) -name "*.go")

FAILPOINT_ENABLE  := $$(find $$PWD/ -type d | grep -vE "(\.git|tools)" | xargs tools/bin/failpoint-ctl enable)
FAILPOINT_DISABLE := $$(find $$PWD/ -type d | grep -vE "(\.git|tools)" | xargs tools/bin/failpoint-ctl disable)

FAIL_ON_STDOUT := awk '{ print } END { if (NR > 0) { exit 1 } }'

LDFLAGS += -s -w
LDFLAGS += -X "github.com/pingcap-inc/tiem/tiup/version.ReleaseVersion=$(shell git describe --tags --dirty --always)"
LDFLAGS += -X "github.com/pingcap-inc/tiem/tiup/version.GitHash=$(shell git rev-parse HEAD)"
LDFLAGS += -X "github.com/pingcap-inc/tiem/tiup/version.GitBranch=$(shell git rev-parse --abbrev-ref HEAD)"

all: fmt build

build: tiup

clean:
	@rm -rf bin

fmt:
	@echo "gofmt (simplify)"
	@gofmt -s -l -w $(FILES) 2>&1 | $(FAIL_ON_STDOUT)

tiup:
	$(GOBUILD) -ldflags '$(LDFLAGS)' -o bin/tiup-em cmd/*.go

check: lint

test: failpoint-enable run-tests failpoint-disable
	@# Target: run tests with failpoint enabled

# run-tests: unit-test integration_test
run-tests: unit-test
	@# Target: run the unit tests

# Run tests
unit-test:
	@# Target: run the code coverage test phase
	mkdir -p cover
	TIUP_HOME=$(shell pwd)/tests/tiup $(GOTEST) ./... -covermode=count -coverprofile cover/cov.unit-test.out

lint: tools/bin/revive
	@# Target: run the lint checker revive
	@echo "linting"
	./tools/check/check-lint.sh
	@tools/bin/revive -formatter friendly -config tools/check/revive.toml $(FILES)

failpoint-enable: tools/bin/failpoint-ctl
	@# Target: enable failpoint
	@$(FAILPOINT_ENABLE)

failpoint-disable: tools/bin/failpoint-ctl
	@# Target: disable failpoint
	@$(FAILPOINT_DISABLE)

tools/bin/failpoint-ctl: go.mod
	@# Target: build the failpoint-ctl utility
	$(GO) build -o $@ github.com/pingcap/failpoint/failpoint-ctl

tools/bin/revive:
	@# Target: build revive utility
	cd tools/check; \
	$(GO) build -o ../bin/revive github.com/mgechev/revive
