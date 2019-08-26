SHELL := /bin/bash

machine.go: machine.go.rl
	ragel -Z -G2 -e -o $@ $<
	@gofmt -w -s $@
	@sed -i '/^\/\/line/d' $@

.PHONY: build
build: machine.go

.PHONY: bench
bench: *_test.go machine.go
	go test -bench=. -benchmem -benchtime=5s ./...

.PHONY: tests
tests: *_test.go machine.go
	go test -race -timeout 10s -coverprofile=coverage.out -covermode=atomic -v ./...