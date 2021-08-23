# Copyright 2019 PingCAP, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

include Makefile.common

.PHONY: all clean test gotest openapi-server cluster-server metadb-server

default: tiupcmd brcmd openapi-server cluster-server metadb-server buildsucc

buildsucc:
	@echo Build TiEM server successfully!

all: dev tiupcmd brcmd openapi-server cluster-server metadb-server

dev: checklist check test

# Install the check tools.
check-setup:bin/revive bin/goword

check: fmt unconvert lint testSuite check-static vet errdoc

fmt:
	@echo "gofmt (simplify)"
	@gofmt -s -l -w $(FILES) 2>&1 | $(FAIL_ON_STDOUT)

goword:bin/goword
	bin/goword $(FILES) 2>&1 | $(FAIL_ON_STDOUT)

check-static: bin/golangci-lint
	bin/golangci-lint run -v $$($(PACKAGE_DIRECTORIES_WITHOUT_BR))

unconvert:bin/unconvert
	@echo "unconvert check"
	@GO111MODULE=on bin/unconvert $(UNCONVERT_PACKAGES)

gogenerate:
	@echo "go generate ./..."
	build_helper/check-gogenerate.sh

errdoc:bin/errdoc-gen
	@echo "generator errors.toml"
	build_helper/check-errdoc.sh

lint:bin/revive
	@echo "linting"
	@bin/revive -formatter friendly -config build_helper/revive.toml $(FILES_WITHOUT_BR)

vet:
	@echo "vet"
	$(GO) vet -all $(PACKAGES_WITHOUT_BR) 2>&1 | $(FAIL_ON_STDOUT)

testSuite:
	@echo "testSuite"
	build_helper/check_testSuite.sh

clean: failpoint-disable
	$(GO) clean -i ./...

# Split tests for CI to run `make test` in parallel.
test: test_part_1
	@>&2 echo "Great, all tests passed."

test_part_1: checklist explaintest

upload-coverage: SHELL:=/bin/bash
upload-coverage:
ifeq ("$(TRAVIS_COVERAGE)", "1")
	mv overalls.coverprofile coverage.txt
	bash <(curl -s https://codecov.io/bash)
endif

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
	curl -s https://codecov.io/bash > $(CODECOV_BASH)
	bash $(CODECOV_BASH) -f $(TEST_DIR)/unit_cov.out -t $(CODECOV_TOKEN)
else
	go tool cover -html "$(TEST_DIR)/all_cov.out" -o "$(TEST_DIR)/all_cov.html"
	go tool cover -html "$(TEST_DIR)/unit_cov.out" -o "$(TEST_DIR)/unit_cov.html"
	go tool cover -func="$(TEST_DIR)/unit_cov.out"
endif

gotest: failpoint-enable
ifeq ("$(TRAVIS_COVERAGE)", "1")
	@echo "Running in TRAVIS_COVERAGE mode."
	$(GO) get github.com/go-playground/overalls
	@export log_level=info; \
	$(OVERALLS) -project=github.com/pingcap/tidb \
			-covermode=count \
			-ignore='.git,br,vendor,cmd,docs,tests,LICENSES' \
			-concurrency=4 \
			-- -coverpkg=./... \
			|| { $(FAILPOINT_DISABLE); exit 1; }
else
	@echo "Running in native mode."
	@export log_level=info; export TZ='Asia/Shanghai'; \
	$(GOTEST) -ldflags '$(TEST_LDFLAGS)' $(EXTRA_TEST_ARGS) -v -cover $(PACKAGES_WITHOUT_BR) -check.p true > gotest.log || { $(FAILPOINT_DISABLE); cat 'gotest.log'; exit 1; }
	@echo "timeout-check"
	grep 'PASS:' gotest.log | go run build_helper/check-timeout.go || { $(FAILPOINT_DISABLE); exit 1; }
endif
	@$(FAILPOINT_DISABLE)

race: failpoint-enable
	@export log_level=debug; \
	$(GOTEST) -timeout 20m -race $(PACKAGES) || { $(FAILPOINT_DISABLE); exit 1; }
	@$(FAILPOINT_DISABLE)

leak: failpoint-enable
	@export log_level=debug; \
	$(GOTEST) -tags leak $(PACKAGES) || { $(FAILPOINT_DISABLE); exit 1; }
	@$(FAILPOINT_DISABLE)

tiupcmd:
	@echo "build tiupcmd start."
ifeq ($(TARGET), "")
	$(GOBUILD) $(RACE_FLAG) -ldflags '$(LDFLAGS) $(CHECK_FLAG)' -o bin/tiupcmd library/secondparty/tiupcmd/main.go
else
	$(GOBUILD) $(RACE_FLAG) -ldflags '$(LDFLAGS) $(CHECK_FLAG)' -o '$(TARGET)' library/secondparty/tiupcmd/main.go
endif
	@echo "build tiupcmd successfully."

brcmd:
	@echo "build brcmd start."
ifeq ($(TARGET), "")
	$(GOBUILD) $(RACE_FLAG) -ldflags '$(LDFLAGS) $(CHECK_FLAG)' -o bin/brcmd library/secondparty/brcmd/main.go
else
	$(GOBUILD) $(RACE_FLAG) -ldflags '$(LDFLAGS) $(CHECK_FLAG)' -o '$(TARGET)' library/secondparty/brcmd/main.go
endif
	@echo "build brcmd successfully."

openapi-server:
	@echo "build openapi-server start."
ifeq ($(TARGET), "")
	$(GOBUILD) $(RACE_FLAG) -ldflags '$(LDFLAGS) $(CHECK_FLAG)' -o bin/openapi-server micro-api/main.go
else
	$(GOBUILD) $(RACE_FLAG) -ldflags '$(LDFLAGS) $(CHECK_FLAG)' -o '$(TARGET)' micro-api/main.go
endif
	@echo "build openapi-server successfully."

cluster-server:
	@echo "build cluster-server start."
ifeq ($(TARGET), "")
	$(GOBUILD) $(RACE_FLAG) -ldflags '$(LDFLAGS) $(CHECK_FLAG)' -o bin/cluster-server micro-cluster/main.go
else
	$(GOBUILD) $(RACE_FLAG) -ldflags '$(LDFLAGS) $(CHECK_FLAG)' -o '$(TARGET)' micro-cluster/main.go
endif
	@echo "build cluster-server successfully."

metadb-server:
	@echo "build metadb-server start."
ifeq ($(TARGET), "")
	$(GOBUILD) $(RACE_FLAG) -ldflags '$(LDFLAGS) $(CHECK_FLAG)' -o bin/metadb-server micro-metadb/main.go
else
	$(GOBUILD) $(RACE_FLAG) -ldflags '$(LDFLAGS) $(CHECK_FLAG)' -o '$(TARGET)' micro-metadb/main.go
endif
	@echo "build metadb-server successfully."

checklist:
	cat checklist.md

failpoint-enable: bin/failpoint-ctl
# Converting gofail failpoints...
	@$(FAILPOINT_ENABLE)

failpoint-disable: bin/failpoint-ctl
# Restoring gofail failpoints...
	@$(FAILPOINT_DISABLE)

bin/megacheck: build_helper/go.mod
	cd build_helper; \
	$(GO) build -o ../bin/megacheck honnef.co/go/build_helper/cmd/megacheck
#TODO

bin/revive: build_helper/go.mod
	cd build_helper; \
	$(GO) build -o ../bin/revive github.com/mgechev/revive

bin/goword: build_helper/go.mod
	cd build_helper; \
	$(GO) build -o ../bin/goword github.com/chzchzchz/goword

bin/unconvert: build_helper/go.mod
	cd build_helper; \
	$(GO) build -o ../bin/unconvert github.com/mdempsky/unconvert

bin/failpoint-ctl: build_helper/go.mod
	cd build_helper; \
	$(GO) build -o ../bin/failpoint-ctl github.com/pingcap/failpoint/failpoint-ctl

bin/errdoc-gen: build_helper/go.mod
	cd build_helper; \
	$(GO) build -o ../bin/errdoc-gen github.com/pingcap/errors/errdoc-gen

bin/golangci-lint:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b ./bin v1.41.1

bin/vfsgendev: build_helper/go.mod
	cd build_helper; \
	$(GO) build -o ../bin/vfsgendev github.com/shurcooL/vfsgen/cmd/vfsgendev

# Usage:

testpkg: failpoint-enable
ifeq ("$(pkg)", "")
	@echo "Require pkg parameter"
else
	@echo "Running unit test for github.com/pingcap/tidb/$(pkg)"
	@export log_level=fatal; export TZ='Asia/Shanghai'; \
	$(GOTEST) -ldflags '$(TEST_LDFLAGS)' -cover github.com/pingcap/tidb/$(pkg) -check.p true -check.timeout 4s || { $(FAILPOINT_DISABLE); exit 1; }
endif
	@$(FAILPOINT_DISABLE)

# Collect the daily benchmark data.
# Usage:
#	make bench-daily TO=/path/to/file.json
#bench-daily:
#	cd ./session && \
#	go test -run TestBenchDaily --date `git log -n1 --date=unix --pretty=format:%cd` --commit `git log -n1 --pretty=format:%h` --outfile $(TO)

#br_web:
#	@cd br/web && npm install && npm run build
#
# There is no FreeBSD environment for GitHub actions. So cross-compile on Linux
# but that doesn't work with CGO_ENABLED=1, so disable cgo. The reason to have
# cgo enabled on regular builds is performance.
#ifeq ("$(GOOS)", "freebsd")
#        GOBUILD  = CGO_ENABLED=0 GO111MODULE=on go build -trimpath -ldflags '$(LDFLAGS)'
#endif
