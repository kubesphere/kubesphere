SHELL := /bin/bash

##@ Format project using goimports tool
format:
	goimports -w .;

##@ Test
test-unit:  ## Run unit tests
	@printf "\033[2m→ Running unit tests...\033[0m\n"
ifdef race
	$(eval testunitargs += "-race")
endif
	$(eval testunitargs += "-cover" "-coverprofile=tmp/unit.cov" "./...")
	@mkdir -p tmp
	@if which gotestsum > /dev/null 2>&1 ; then \
		echo "gotestsum --format=short-verbose --junitfile=tmp/unit-report.xml --" $(testunitargs); \
		gotestsum --format=short-verbose --junitfile=tmp/unit-report.xml -- $(testunitargs); \
	else \
		echo "go test -v" $(testunitargs); \
		go test -v $(testunitargs); \
	fi;
test: test-unit

test-integ:  ## Run integration tests
	@printf "\033[2m→ Running integration tests...\033[0m\n"
	$(eval testintegtags += "integration")
ifdef multinode
	$(eval testintegtags += "multinode")
endif
ifdef race
	$(eval testintegargs += "-race")
endif
	$(eval testintegargs += "-cover" "-coverprofile=tmp/integration-client.cov" "-tags='$(testintegtags)'" "-timeout=1h")
	@mkdir -p tmp
	@if which gotestsum > /dev/null 2>&1 ; then \
		echo "gotestsum --format=short-verbose --junitfile=tmp/integration-report.xml --" $(testintegargs); \
		gotestsum --format=short-verbose --junitfile=tmp/integration-report.xml -- $(testintegargs) "."; \
		gotestsum --format=short-verbose --junitfile=tmp/integration-report.xml -- $(testintegargs) "./opensearchtransport" "./opensearchapi" "./opensearchutil"; \
	else \
		echo "go test -v" $(testintegargs) "."; \
		go test -v $(testintegargs) "./opensearchtransport" "./opensearchapi" "./opensearchutil"; \
	fi;

test-integ-secure: ##Run secure integration tests
	go test -tags=secure,integration ./opensearch_secure_integration_test.go

test-bench:  ## Run benchmarks
	@printf "\033[2m→ Running benchmarks...\033[0m\n"
	go test -run=none -bench=. -benchmem ./...

test-coverage:  ## Generate test coverage report
	@printf "\033[2m→ Generating test coverage report...\033[0m\n"
	@go tool cover -html=tmp/unit.cov -o tmp/coverage.html
	@go tool cover -func=tmp/unit.cov | 'grep' -v 'opensearchapi/api\.' | sed 's/github.com\/opensearch-project\/opensearch-go\///g'
	@printf "\033[0m--------------------------------------------------------------------------------\nopen tmp/coverage.html\n\n\033[0m"

##@ Development
lint:  ## Run lint on the package
	@printf "\033[2m→ Running lint...\033[0m\n"
	go vet github.com/opensearch-project/opensearch-go/...
	go list github.com/opensearch-project/opensearch-go/... | 'grep' -v internal | xargs golint -set_exit_status
	@{ \
		set -e ; \
		trap "test -d ../../../.git && git checkout --quiet go.mod" INT TERM EXIT; \
		echo "cd internal/build/ && go vet ./..."; \
		cd "internal/build/" && go mod tidy && go mod download && go vet ./...; \
	}


backport: ## Backport one or more commits from master into version branches
ifeq ($(origin commits), undefined)
	@echo "Missing commit(s), exiting..."
	@exit 2
endif
ifndef branches
	$(eval branches_list = '1.x')
else
	$(eval branches_list = $(shell echo $(branches) | tr ',' ' ') )
endif
	$(eval commits_list = $(shell echo $(commits) | tr ',' ' '))
	@printf "\033[2m→ Backporting commits [$(commits)]\033[0m\n"
	@{ \
		set -e -o pipefail; \
		for commit in $(commits_list); do \
			git show --pretty='%h | %s' --no-patch $$commit; \
		done; \
		echo ""; \
		for branch in $(branches_list); do \
			printf "\033[2m→ $$branch\033[0m\n"; \
			git checkout $$branch; \
			for commit in $(commits_list); do \
				git cherry-pick -x $$commit; \
			done; \
			git status --short --branch; \
			echo ""; \
		done; \
		printf "\033[2m→ Push updates to Github:\033[0m\n"; \
		for branch in $(branches_list); do \
			echo "git push --verbose origin $$branch"; \
		done; \
	}

release: ## Release a new version to Github
	$(eval branch = $(shell git rev-parse --abbrev-ref HEAD))
	$(eval current_version = $(shell cat internal/version/version.go | sed -Ee 's/const Client = "(.*)"/\1/' | tail -1))
	@printf "\033[2m→ [$(branch)] Current version: $(current_version)...\033[0m\n"
ifndef version
	@printf "\033[31m[!] Missing version argument, exiting...\033[0m\n"
	@exit 2
endif
ifeq ($(version), "")
	@printf "\033[31m[!] Empty version argument, exiting...\033[0m\n"
	@exit 2
endif
	@printf "\033[2m→ [$(branch)] Creating version $(version)...\033[0m\n"
	@{ \
		set -e -o pipefail; \
		cp internal/version/version.go internal/version/version.go.OLD && \
		cat internal/version/version.go.OLD | sed -e 's/Client = ".*"/Client = "$(version)"/' > internal/version/version.go && \
		go vet internal/version/version.go && \
		go fmt internal/version/version.go && \
		git diff --color-words internal/version/version.go | tail -n 1; \
	}
	@{ \
		set -e -o pipefail; \
		printf "\033[2m→ Commit and create Git tag? (y/n): \033[0m\c"; \
		read continue; \
		if [[ $$continue == "y" ]]; then \
			git add internal/version/version.go && \
			git commit --no-status --quiet --message "Release $(version)" && \
			git tag --annotate v$(version) --message 'Release $(version)'; \
			printf "\033[2m→ Push `git show --pretty='%h (%s)' --no-patch HEAD` to Github:\033[0m\n\n"; \
			printf "\033[1m  git push origin HEAD && git push origin v$(version)\033[0m\n\n"; \
			mv internal/version/version.go.OLD internal/version/version.go && \
			git add internal/version/version.go && \
			original_version=`cat internal/version/version.go | sed -ne 's;^const Client = "\(.*\)"$$;\1;p'` && \
			git commit --no-status --quiet --message "Update version to $$original_version"; \
			printf "\033[2m→ Version updated to [$$original_version].\033[0m\n\n"; \
		else \
			echo "Aborting..."; \
			rm internal/version/version.go.OLD; \
			exit 1; \
		fi; \
	}

godoc: ## Display documentation for the package
	@printf "\033[2m→ Generating documentation...\033[0m\n"
	@echo "* http://localhost:6060/pkg/github.com/opensearch-project/opensearch-go"
	@echo "* http://localhost:6060/pkg/github.com/opensearch-project/opensearch-go/opensearchapi"
	@echo "* http://localhost:6060/pkg/github.com/opensearch-project/opensearch-go/opensearchtransport"
	@echo "* http://localhost:6060/pkg/github.com/opensearch-project/opensearch-go/opensearchutil"
	@printf "\n"
	godoc --http=localhost:6060 --play

cluster.opendistro.build:
	docker-compose --project-directory .ci/opendistro build;

cluster.opendistro.start:
	docker-compose --project-directory .ci/opendistro up -d ;
	sleep 20;

cluster.opendistro.stop:
	docker-compose --project-directory .ci/opendistro down ;

cluster.opensearch.build:
	docker-compose --project-directory .ci/opensearch build;

cluster.opensearch.start:
	docker-compose --project-directory .ci/opensearch up -d ;
	sleep 20;

cluster.opensearch.stop:
	docker-compose --project-directory .ci/opensearch down ;


cluster.clean: ## Remove unused Docker volumes and networks
	@printf "\033[2m→ Cleaning up Docker assets...\033[0m\n"
	docker volume prune --force
	docker network prune --force
	docker system prune --volumes --force




workflow: ## Run all github workflow commands here sequentially

# Lint
	make lint
# License Checker
	.github/check-license-headers.sh
# Unit Test
	make test-unit race=true
# Benchmarks Test
	make test-bench
# Integration Test
### OpenDistro
	make cluster.clean cluster.opendistro.build cluster.opendistro.start
	make test-integ race=true
	make cluster.opendistro.stop
### OpenSearch
	make cluster.clean cluster.opensearch.build cluster.opensearch.start
	make test-integ race=true
	make cluster.opensearch.stop

##@ Other
#------------------------------------------------------------------------------
help:  ## Display help
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
#------------- <https://suva.sh/posts/well-documented-makefiles> --------------

.DEFAULT_GOAL := help
.PHONY: help backport cluster cluster.opendistro.build cluster.opendistro.start cluster.opendistro.stop cluster.clean coverage  godoc lint release test test-bench test-integ test-unit
