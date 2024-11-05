all: test install

install:
	go install

lint:
	gofmt -l -s -w . && go vet .

test:
	go test -race -v -coverprofile=coverage.txt -covermode=atomic

bench:
	go test -run=XXX -bench=. -benchmem -count=5
