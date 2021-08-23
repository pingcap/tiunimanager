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

TIEM_BINARY_DIR = ${CURDIR}/bin
TIUPCMD_BINARY = ${TIEM_BINARY_DIR}/tiupcmd
BRCMD_BINARY = ${TIEM_BINARY_DIR}/brcmd
OPENAPI_SERVER_BINARY = ${TIEM_BINARY_DIR}/openapi-server
CLUSTER_SERVER_BINARY = ${TIEM_BINARY_DIR}/cluster-server
METADB_SERVER_BINARY = ${TIEM_BINARY_DIR}/metadb-server
TIEM_INSTALL_PREFIX = ${PREFIX}/tiem

include Makefile.common

.PHONY: all clean test gotest gotool help
all:
	build

# 1. build binary
build:
	@echo "build TiEM server start."
	make build_tiupcmd
	make build_brcmd
	make build_openapi_server
	make build_cluster_server
	make build_metadb_server
	@echo "build TiEM all server successful."

#Compile all TiEM microservices
build_tiupcmd:
	@echo "build tiupcmd start."
	$(GOBUILD) $(RACE_FLAG) -ldflags '$(LDFLAGS) $(CHECK_FLAG)' -o ${TIUPCMD_BINARY} library/secondparty/tiupcmd/main.go
	@echo "build tiupcmd sucessufully."

build_brcmd:
	@echo "build brcmd start."
	$(GOBUILD) $(RACE_FLAG) -ldflags '$(LDFLAGS) $(CHECK_FLAG)' -o ${BRCMD_BINARY} library/secondparty/brcmd/main.go
	@echo "build brcmd sucessufully."

build_openapi_server:
	@echo "build openapi-server start."
	$(GOBUILD) $(RACE_FLAG) -ldflags '$(LDFLAGS) $(CHECK_FLAG)' -o ${OPENAPI_SERVER_BINARY} micro-api/main.go
	@echo "build openapi-server sucessufully."

build_cluster_server:
	@echo "build cluster-server start."
	$(GOBUILD) $(RACE_FLAG) -ldflags '$(LDFLAGS) $(CHECK_FLAG)' -o ${CLUSTER_SERVER_BINARY} micro-cluster/main.go
	@echo "build cluster-server sucessufully."

build_metadb_server:
	@echo "build metadb-server start."
	$(GOBUILD) $(RACE_FLAG) -ldflags '$(LDFLAGS) $(CHECK_FLAG)' -o ${METADB_SERVER_BINARY} micro-metadb/main.go
	@echo "build metadb-server sucessufully."

#2. R&D to test the code themselves for compliance before submitting it
devselfcheck:
	cat prchecklist.md
	make gotool
	@echo "start self check."
	make check_fmt
	make check_goword
	make check_static
	make check_unconvert
	make check_lint
	make check_vet
	@echo "self check complete."

gotool:
	@echo "build compilation toolchain start."
	make build_revive
	make build_goword
	make build_unconvert
	make build_failpoint_ctl
	make build_errdoc_gen
	make build_golangci_lint
	make build_vfsgendev
	@echo "build compilation toolchain successful."

#Get and compile the tools required in the project
build_revive: build_helper/go.mod
	cd build_helper; \
	$(GO) build -o ${TIEM_BINARY_DIR}/revive github.com/mgechev/revive

build_goword: build_helper/go.mod
	cd build_helper; \
	$(GO) build -o ${TIEM_BINARY_DIR}/goword github.com/chzchzchz/goword

build_unconvert: build_helper/go.mod
	cd build_helper; \
	$(GO) build -o ${TIEM_BINARY_DIR}/unconvert github.com/mdempsky/unconvert

build_failpoint_ctl: build_helper/go.mod
	cd build_helper; \
	$(GO) build -o ${TIEM_BINARY_DIR}/failpoint-ctl github.com/pingcap/failpoint/failpoint-ctl

build_errdoc_gen: build_helper/go.mod
	cd build_helper; \
	$(GO) build -o ${TIEM_BINARY_DIR}/errdoc-gen github.com/pingcap/errors/errdoc-gen

build_golangci_lint:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b ${TIEM_BINARY_DIR} v1.41.1

build_vfsgendev: build_helper/go.mod
	cd build_helper; \
	$(GO) build -o ${TIEM_BINARY_DIR}/vfsgendev github.com/shurcooL/vfsgen/cmd/vfsgendev

unconvert:bin/unconvert
	@echo "unconvert check"
	@GO111MODULE=on bin/unconvert $(UNCONVERT_PACKAGES)

build_megacheck: build_helper/go.mod
	cd build_helper; \
	$(GO) build -o ${TIEM_BINARY_DIR}/megacheck honnef.co/go/build_helper/cmd/megacheck

check_fmt:
	@echo "gofmt (simplify)"
	#@gofmt -s -l -w $(FILES) 2>&1 | $(FAIL_ON_STDOUT)

check_goword:
	@echo "goword check, files: ${FILES}"
	#${TIEM_BINARY_DIR}/goword $(FILES) 2>&1 | $(FAIL_ON_STDOUT)

check_static:
	@echo "code static check, files: $($(PACKAGE_LIST))"
	#${TIEM_BINARY_DIR}/golangci-lint run -v $$($(PACKAGE_DIRECTORIES))

check_unconvert:
	@echo "unconvert check, files: $($(PACKAGE_LIST))"
	#@GO111MODULE=on ${TIEM_BINARY_DIR}/unconvert $(UNCONVERT_PACKAGES)

#TODO
#check_gogenerate:
#	@echo "go generate check."
#	build_helper/check-gogenerate.sh
#
#check_errdoc:
#	@echo "error docs check"
#	build_helper/check-errdoc.sh

check_lint:
	@echo "linting check"
	#@${TIEM_BINARY_DIR}/revive -formatter friendly -config build_helper/revive.toml $(FILES_WITHOUT_BR)

check_vet:
	@echo "vet check"
	#$(GO) vet -all $(PACKAGES_WITHOUT_BR) 2>&1 | $(FAIL_ON_STDOUT)

#TODO
#check_testsuite:
#	@echo "testsuite check"
#	build_helper/check_testSuite.sh

install:
	mkdir -p ${TIEM_INSTALL_PREFIX}
	mkdir -p ${TIEM_INSTALL_PREFIX}/bin
	mkdir -p ${TIEM_INSTALL_PREFIX}/etc
	mkdir -p ${TIEM_INSTALL_PREFIX}/logs
	mkdir -p ${TIEM_INSTALL_PREFIX}/scripts
	mkdir -p ${TIEM_INSTALL_PREFIX}/docs
	cp ${TIUPCMD_BINARY} ${TIEM_INSTALL_PREFIX}/bin
	cp ${BRCMD_BINARY} ${TIEM_INSTALL_PREFIX}/bin
	cp ${OPENAPI_SERVER_BINARY} ${TIEM_INSTALL_PREFIX}/bin
	cp ${METADB_SERVER_BINARY} ${TIEM_INSTALL_PREFIX}/bin
	cp ${CLUSTER_SERVER_BINARY} ${TIEM_INSTALL_PREFIX}/bin

uninstall:
	@if [ -d ${TIEM_INSTALL_PREFIX} ] ; then rm ${TIEM_INSTALL_PREFIX} ; fi

clean:
	@if [ -f ${BRCMD_BINARY} ] ; then rm ${BRCMD_BINARY} ; fi
	@if [ -f ${TIUPCMD_BINARY} ] ; then rm ${TIUPCMD_BINARY} ; fi
	@if [ -f ${OPENAPI_SERVER_BINARY} ] ; then rm ${OPENAPI_SERVER_BINARY} ; fi
	@if [ -f ${CLUSTER_SERVER_BINARY} ] ; then rm ${CLUSTER_SERVER_BINARY} ; fi
	@if [ -f ${METADB_SERVER_BINARY} ] ; then rm ${METADB_SERVER_BINARY} ; fi
	@if [ -f ${TIEM_BINARY_DIR}/revive ] ; then rm ${TIEM_BINARY_DIR}/revive ; fi
	@if [ -f ${TIEM_BINARY_DIR}/goword ] ; then rm ${TIEM_BINARY_DIR}/goword ; fi
	@if [ -f ${TIEM_BINARY_DIR}/unconvert ] ; then rm ${TIEM_BINARY_DIR}/unconvert ; fi
	@if [ -f ${TIEM_BINARY_DIR}/failpoint-ctl ] ; then rm ${TIEM_BINARY_DIR}/failpoint-ctl; fi
	@if [ -f ${TIEM_BINARY_DIR}/vfsgendev ] ; then rm ${TIEM_BINARY_DIR}/vfsgendev; fi
	@if [ -f ${TIEM_BINARY_DIR}/golangci-lint ] ; then rm ${TIEM_BINARY_DIR}/golangci-lint; fi

help:
	@echo "make build, build binary for all servers"
	@echo "make install, install binary to target "
	@echo "make uninstall, uninstall binary from target "
	@echo "make devselfcheck, dev commit code precheck."
	@echo "make test, test all case"
	@echo "make upload_coverage, upload coverage information"

# upload coverage information
upload_coverage: SHELL:=/bin/bash
upload_coverage:
ifeq ("$(TRAVIS_COVERAGE)", "1")
	mv overalls.coverprofile coverage.txt
	bash <(curl -s https://codecov.io/bash)
endif

test:
	make failpoint-enable

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
	$(OVERALLS) -project=github.com/pingcap/tiem\
			-covermode=count \
			-ignore='.git,vendor,cmd,docs,tests,LICENSES' \
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

#race: failpoint-enable
#   @export log_level=debug; \
#	$(GOTEST) -timeout 20m -race $(PACKAGES) || { $(FAILPOINT_DISABLE); exit 1; }
#	@$(FAILPOINT_DISABLE)
#
#leak: failpoint-enable
#	@export log_level=debug; \
#	$(GOTEST) -tags leak $(PACKAGES) || { $(FAILPOINT_DISABLE); exit 1; }
#	@$(FAILPOINT_DISABLE)
#
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

#checklist:
#	cat prchecklist.md
#
#failpoint-enable: bin/failpoint-ctl
## Converting gofail failpoints...
#	@$(FAILPOINT_ENABLE)
#
#failpoint-disable: bin/failpoint-ctl
## Restoring gofail failpoints...
#	@$(FAILPOINT_DISABLE)
#
##TODO
#

## Usage:
#
#testpkg: failpoint-enable
#ifeq ("$(pkg)", "")
#	@echo "Require pkg parameter"
#else
#	@echo "Running unit test for github.com/pingcap/tidb/$(pkg)"
#	@export log_level=fatal; export TZ='Asia/Shanghai'; \
#	$(GOTEST) -ldflags '$(TEST_LDFLAGS)' -cover github.com/pingcap/tidb/$(pkg) -check.p true -check.timeout 4s || { $(FAILPOINT_DISABLE); exit 1; }
#endif
#	@$(FAILPOINT_DISABLE)
#
## Collect the daily benchmark data.
## Usage:
##	make bench-daily TO=/path/to/file.json
##bench-daily:
##	cd ./session && \
##	go test -run TestBenchDaily --date `git log -n1 --date=unix --pretty=format:%cd` --commit `git log -n1 --pretty=format:%h` --outfile $(TO)
