.PHONY: all build install clean test help

path_to_add := $(addsuffix /bin,$(subst :,/bin:,$(GOPATH)))
export PATH := $(path_to_add):$(PATH)

TEST_DIR := /tmp/tidb_em_test

GO       := GO111MODULE=on go

GOVERSIONGE116 := $(shell expr $$(go version|cut -f3 -d' '|tr -d "go"|cut -f2 -d.) \>= 16)

ifeq ($(GOVERSION114), 1)
GOTEST   := CGO_ENABLED=1 $(GO) test -p 3 --race -gcflags=all=-d=checkptr=0
else
GOTEST   := CGO_ENABLED=1 $(GO) test -p 3 --race
endif

PACKAGE_LIST := go list ./...| grep -vE 'docs|proto'
PACKAGES  := $$($(PACKAGE_LIST))
EM_PKG := github.com/pingcap/tiem

all: build
build:
	@bash make.sh build
install:
	@bash make.sh install $(INSTALL_DIR)
clean:
	@bash make.sh clean
test: unit_test
help:
	@echo "make build"
	@echo "make INSTALL_DIR=installDir install"
	@echo "make clean"
	@echo "make test"
unit_test:
	mkdir -p "$(TEST_DIR)"
	$(GOTEST) -cover -covermode=atomic -coverprofile="$(TEST_DIR)/cov.unit.out" $(PACKAGES)

coverage:
#	GO111MODULE=off go get github.com/wadey/gocovmerge
#	gocovmerge "$(TEST_DIR)"/cov.* | grep -vE ".*.pb.go" > "$(TEST_DIR)/all_cov.out"
	grep -vE ".*.pb.go" "$(TEST_DIR)/cov.unit.out" > "$(TEST_DIR)/unit_cov.out"
ifeq ("$(JenkinsCI)", "1")
#	GO111MODULE=off go get github.com/mattn/goveralls
#	@goveralls -coverprofile=$(TEST_DIR)/all_cov.out -service=jenkins-ci -repotoken $(COVERALLS_TOKEN)
	bash <(curl -s https://codecov.io/bash) -f $(TEST_DIR)/unit_cov.out -t $(CODECOV_TOKEN)
else
	go tool cover -html "$(TEST_DIR)/all_cov.out" -o "$(TEST_DIR)/all_cov.html"
	go tool cover -html "$(TEST_DIR)/unit_cov.out" -o "$(TEST_DIR)/unit_cov.html"
	go tool cover -func="$(TEST_DIR)/unit_cov.out"
endif