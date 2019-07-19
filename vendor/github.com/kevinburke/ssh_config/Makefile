BUMP_VERSION := $(GOPATH)/bin/bump_version
MEGACHECK := $(GOPATH)/bin/megacheck
WRITE_MAILMAP := $(GOPATH)/bin/write_mailmap

IGNORES := 'github.com/kevinburke/ssh_config/config.go:U1000 github.com/kevinburke/ssh_config/config.go:S1002 github.com/kevinburke/ssh_config/token.go:U1000'

$(MEGACHECK):
	go get honnef.co/go/tools/cmd/megacheck

lint: $(MEGACHECK)
	go vet ./...
	$(MEGACHECK) --ignore=$(IGNORES) ./...

test: lint
	@# the timeout helps guard against infinite recursion
	go test -timeout=250ms ./...

race-test: lint
	go test -timeout=500ms -race ./...

$(BUMP_VERSION):
	go get -u github.com/kevinburke/bump_version

release: test | $(BUMP_VERSION)
	$(BUMP_VERSION) minor config.go

force: ;

AUTHORS.txt: force | $(WRITE_MAILMAP)
	$(WRITE_MAILMAP) > AUTHORS.txt

authors: AUTHORS.txt
