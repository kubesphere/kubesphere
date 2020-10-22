all: test

test:
	go vet .
	go test -cover -v .

ex:
	cd examples && ls *.go | xargs go build -o /tmp/ignore