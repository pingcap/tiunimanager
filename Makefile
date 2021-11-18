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
FILE_SERVER_BINARY = ${TIEM_BINARY_DIR}/file-server
TIEM_INSTALL_PREFIX = ${PREFIX}/tiem
PROTOC_GEN_MICRO = github.com/asim/go-micro/cmd/protoc-gen-micro/v3@v3.7.0
PROTOC_GEN_GO = google.golang.org/protobuf/cmd/protoc-gen-go@v1.27.1
PROTOBUF_VERSION = 3.14.0
PROTOC_PKG = protoc-$(PROTOBUF_VERSION)-$(OS)-$(ARCH).zip
PROTOC_URL = https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOBUF_VERSION)/$(PROTOC_PKG)
GENERATE_TARGET_DIR = $(CURDIR)/library/client/

include Makefile.common

.PHONY: all clean test gotest gotool help proto
all: prepare proto build

# download protoc-gen-micro, protoc-gen-go and protoc
prepare:
	@if [ ! -f "$(GOPATH)/bin/protoc-gen-micro" ]; then echo "download $(PROTOC_GEN_MICRO)"; fi
	@if [ ! -f "$(GOPATH)/bin/protoc-gen-micro" ]; then $(GO) install $(PROTOC_GEN_MICRO); fi
	@if [ ! -f "$(GOPATH)/bin/protoc-gen-go" ]; then echo "download $(PROTOC_GEN_GO)"; fi
	@if [ ! -f "$(GOPATH)/bin/protoc-gen-go" ]; then $(GO) install $(PROTOC_GEN_GO); fi

	@if [ ! -f "$(GOPATH)/bin/protoc" ]; then echo "download $(PROTOC_URL)"; fi
	@if [ ! -f "$(GOPATH)/bin/protoc" ]; then curl -OL $(PROTOC_URL); fi
	@if [ -f "$(CURDIR)/$(PROTOC_PKG)" ]; then unzip -o $(PROTOC_PKG) -d $(GOPATH) bin/protoc; fi
	@if [ -f "$(CURDIR)/$(PROTOC_PKG)" ]; then unzip -o $(PROTOC_PKG) -d $(GOPATH) 'include/*'; fi
	@if [ -f "$(CURDIR)/$(PROTOC_PKG)" ]; then rm -rf $(PROTOC_PKG); fi

# generate protobuf files
proto:
	@echo "start to generate protobuf files"
	@if [ ! -d "$(GENERATE_TARGET_DIR)/cluster" ]; then mkdir -p $(GENERATE_TARGET_DIR)/cluster; fi
	@if [ ! -d "$(GENERATE_TARGET_DIR)/metadb" ]; then mkdir -p $(GENERATE_TARGET_DIR)/metadb; fi
	protoc --proto_path=$(CURDIR)/proto/cluster:$(GOPATH)/include --micro_out=$(GENERATE_TARGET_DIR)/cluster \
		--go_out=$(GENERATE_TARGET_DIR)/cluster $(CURDIR)/proto/cluster/*.proto
	protoc --proto_path=$(CURDIR)/proto/metadb:$(GOPATH)/include --micro_out=$(GENERATE_TARGET_DIR)/metadb \
		--go_out=$(GENERATE_TARGET_DIR)/metadb $(CURDIR)/proto/metadb/*.proto
	@echo "generate protobuf files successfully"


# 1. build binary
build:
	@echo "build TiEM server start."
	make build_openapi_server
	make build_cluster_server
	make build_metadb_server
	make build_file_server
	@echo "build TiEM all server successfully."

#Compile all TiEM microservices
build_openapi_server:
	@echo "build openapi-server start."
	$(GOBUILD) $(RACE_FLAG) -ldflags '$(LDFLAGS) $(CHECK_FLAG)' -o ${OPENAPI_SERVER_BINARY} micro-api/*.go
	@echo "build openapi-server successfully."

build_cluster_server:
	@echo "build cluster-server start."
	$(GOBUILD) $(RACE_FLAG) -ldflags '$(LDFLAGS) $(CHECK_FLAG)' -o ${CLUSTER_SERVER_BINARY} micro-cluster/*.go
	@echo "build cluster-server successfully."

build_metadb_server:
	@echo "build metadb-server start."
	$(GOBUILD) $(RACE_FLAG) -ldflags '$(LDFLAGS) $(CHECK_FLAG)' -o ${METADB_SERVER_BINARY} micro-metadb/*.go
	@echo "build metadb-server successfully."

build_file_server:
	@echo "build file-server start."
	$(GOBUILD) $(RACE_FLAG) -ldflags '$(LDFLAGS) $(CHECK_FLAG)' -o ${FILE_SERVER_BINARY} file-server/*.go
	@echo "build file-server sucessufully."

#2. R&D to test the code themselves for compliance before submitting it
devselfcheck:
	cat resource/prchecklist.md
	make gotool
	@echo "start self check."
	make check_fmt
	make check_goword
	make check_static
	make check_unconvert
	make check_lint
	make check_vet
	make test
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
	@echo "build compilation toolchain successfully."

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

check_lint:
	@echo "linting check"
	#@${TIEM_BINARY_DIR}/revive -formatter friendly -config build_helper/revive.toml $(FILES_WITHOUT_BR)

check_vet:
	@echo "vet check"
	#$(GO) vet -all $(PACKAGES_WITHOUT_BR) 2>&1 | $(FAIL_ON_STDOUT)

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
	cp ${FILE_SERVER_BINARY} ${TIEM_INSTALL_PREFIX}/bin

uninstall:
	@echo "uninstall: remove all files in $(TIEM_INSTALL_PREFIX)"
	@if [ -d ${TIEM_INSTALL_PREFIX} ] ; then rm ${TIEM_INSTALL_PREFIX} ; fi

clean:
	@if [ -f ${BRCMD_BINARY} ] ; then rm ${BRCMD_BINARY} ; fi
	@if [ -f ${TIUPCMD_BINARY} ] ; then rm ${TIUPCMD_BINARY} ; fi
	@if [ -f ${OPENAPI_SERVER_BINARY} ] ; then rm ${OPENAPI_SERVER_BINARY} ; fi
	@if [ -f ${CLUSTER_SERVER_BINARY} ] ; then rm ${CLUSTER_SERVER_BINARY} ; fi
	@if [ -f ${METADB_SERVER_BINARY} ] ; then rm ${METADB_SERVER_BINARY} ; fi
	@if [ -f ${FILE_SERVER_BINARY} ] ; then rm ${FILE_SERVER_BINARY} ; fi
	@if [ -f ${TIEM_BINARY_DIR}/revive ] ; then rm ${TIEM_BINARY_DIR}/revive ; fi
	@if [ -f ${TIEM_BINARY_DIR}/goword ] ; then rm ${TIEM_BINARY_DIR}/goword ; fi
	@if [ -f ${TIEM_BINARY_DIR}/unconvert ] ; then rm ${TIEM_BINARY_DIR}/unconvert ; fi
	@if [ -f ${TIEM_BINARY_DIR}/failpoint-ctl ] ; then rm ${TIEM_BINARY_DIR}/failpoint-ctl; fi
	@if [ -f ${TIEM_BINARY_DIR}/vfsgendev ] ; then rm ${TIEM_BINARY_DIR}/vfsgendev; fi
	@if [ -f ${TIEM_BINARY_DIR}/golangci-lint ] ; then rm ${TIEM_BINARY_DIR}/golangci-lint; fi
	@if [ -f ${TIEM_BINARY_DIR}/errdoc-gen ] ; then rm ${TIEM_BINARY_DIR}/errdoc-gen; fi
	@if [ -d ${GENERATE_TARGET_DIR}/cluster ] ; then rm -rf ${GENERATE_TARGET_DIR}/cluster; fi
	@if [ -d ${GENERATE_TARGET_DIR}/metadb ] ; then rm -rf ${GENERATE_TARGET_DIR}/metadb; fi
	@if [ -d ${CURDIR}/test ] ; then rm -rf ${CURDIR}/test; fi

help:
	@echo "make build, build binary for all servers"
	@echo "make install, install binary to target "
	@echo "make uninstall, uninstall binary from target "
	@echo "make devselfcheck, dev commit code precheck."
	@echo "make test, test all case"
	@echo "make upload_coverage, upload coverage information"

upload_coverage:
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

# don't run it locally, only for CI, use test instead
ci_test: add_test_file prepare proto build mock
	GO111MODULE=off go get github.com/axw/gocov/gocov
	GO111MODULE=off go get github.com/jstemmer/go-junit-report
	GO111MODULE=off go get github.com/AlekSi/gocov-xml
	go test -v ${PACKAGES} -coverprofile=cover.out |go-junit-report > test.xml
	gocov convert cover.out | gocov-xml > coverage.xml

test: prepare proto build mock
	mkdir -p "$(TEST_DIR)"
	-go test -v ${PACKAGES} -coverprofile="$(TEST_DIR)/cover.out"
	go tool cover -html "$(TEST_DIR)/cover.out" -o "$(TEST_DIR)/cover.html"
	echo "check coverage info by opening $(TEST_DIR)/cover.html through browser"

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

failpoint-enable: build_failpoint_ctl
# Converting gofail failpoints...
	@$(FAILPOINT_ENABLE)

failpoint-disable: build_failpoint_ctl
# Restoring gofail failpoints...
	@$(FAILPOINT_DISABLE)

lint:
	# refer https://golangci-lint.run/usage/install/#local-installation to install golangci-lint firstly
	-golangci-lint run --out-format=junit-xml  --timeout=10m -v ./... > golangci-lint-report.xml

gosec:
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec -fmt=junit-xml -out=results.xml -stdout -verbose=text -exclude=G103,G104,G204,G304,G307,G401,G404,G501,G505,G601 ./... || true

add_test_file:
	build_helper/add_test_file.sh

mock:
	$(GO) install github.com/golang/mock/mockgen
	mockgen -destination ./test/mocksecondparty/mock_second_party_manager.go -package mocksecondparty -source ./library/secondparty/second_party_manager.go
	mockgen -destination ./test/mockdb/mock_db.pb.micro.go -package mockdb -source ./library/client/metadb/dbpb/db.pb.micro.go
	mockgen -destination ./test/mockcluster/mock_cluster.pb.micro.go -package mockcluster -source ./library/client/cluster/clusterpb/cluster.pb.micro.go


