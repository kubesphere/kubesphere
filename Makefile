# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

TRAG.Org:=kubesphere
TRAG.Name:=ks-apiserver
TRAG.Gopkg:=kubesphere.io/kubesphere
TRAG.Version:=$(TRAG.Gopkg)/pkg/version

DOCKER_TAGS=latest
RUN_IN_DOCKER:=docker run -it --rm -v `pwd`:/go/src/$(TRAG.Gopkg) -v `pwd`/tmp/cache:/root/.cache/go-build  -w /go/src/$(TRAG.Gopkg) -e GOBIN=/go/src/$(TRAG.Gopkg)/tmp/bin -e USER_ID=`id -u` -e GROUP_ID=`id -g` kubesphere/kubesphere-builder
GO_FMT:=goimports -l -w -e -local=kubesphere -srcdir=/go/src/$(TRAG.Gopkg)
GO_FILES:=./cmd ./pkg

define get_diff_files
    $(eval DIFF_FILES=$(shell git diff --name-only --diff-filter=ad | grep -E "^(test|cmd|pkg)/.+\.go"))
endef
define get_build_flags
    $(eval SHORT_VERSION=$(shell git describe --tags --always --dirty="-dev"))
    $(eval SHA1_VERSION=$(shell git show --quiet --pretty=format:%H))
	$(eval DATE=$(shell date +'%Y-%m-%dT%H:%M:%S'))
	$(eval BUILD_FLAG= -X $(TRAG.Version).ShortVersion="$(SHORT_VERSION)" \
		-X $(TRAG.Version).GitSha1Version="$(SHA1_VERSION)" \
		-X $(TRAG.Version).BuildDate="$(DATE)")
endef

.PHONY: all
all: generate build

.PHONY: help
help:
# TODO: update help info to last version
	@echo "TODO"

.PHONY: init-vendor
init-vendor:
	@if [[ ! -f "$$(which govendor)" ]]; then \
		go get -u github.com/kardianos/govendor; \
	fi
	govendor init
	govendor add +external
	@echo "init-vendor done"

.PHONY: update-vendor
update-vendor:
	@if [[ ! -f "$$(which govendor)" ]]; then \
		go get -u github.com/kardianos/govendor; \
	fi
	govendor update +external
	govendor list
	@echo "update-vendor done"

.PHONY: update-builder
update-builder:
	docker pull kubesphere/kubesphere-builder
	@echo "update-builder done"

.PHONY: generate-in-local
generate-in-local:
	go generate ./pkg/version/

.PHONY: generate
generate:
	$(RUN_IN_DOCKER) make generate-in-local
	@echo "generate done"

.PHONY: fmt-all
fmt-all:
	mkdir -p ./tmp/bin && cp -r ./install ./tmp/
	$(RUN_IN_DOCKER) $(GO_FMT) $(GO_FILES)
	@echo "fmt done"

.PHONY: fmt
fmt:
	$(call get_diff_files)
	$(if $(DIFF_FILES), \
		$(RUN_IN_DOCKER) $(GO_FMT) ${DIFF_FILES}, \
		$(info cannot find modified files from git) \
	)
	@echo "fmt done"

.PHONY: fmt-check
fmt-check: fmt-all
	$(call get_diff_files)
	$(if $(DIFF_FILES), \
		exit 2 \
	)

.PHONY: build
build: fmt
	mkdir -p ./tmp/bin && cp -r ./install/ ./tmp/
	$(call get_build_flags)
	$(RUN_IN_DOCKER) time go install -ldflags '$(BUILD_FLAG)' $(TRAG.Gopkg)/cmd/...
	mv ./tmp/bin/cmd ./tmp/bin/$(TRAG.Name)
	@docker build -t $(TRAG.Org)/$(TRAG.Name) -f ./Dockerfile.dev ./tmp
	@docker image prune -f 1>/dev/null 2>&1
	@echo "build done"

.PHONY: release
release:
	@echo "TODO"

.PHONY: clean
clean:
	-make -C ./pkg/version clean
	@echo "ok"
