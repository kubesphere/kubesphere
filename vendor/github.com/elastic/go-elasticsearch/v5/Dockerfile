# $ docker build --file Dockerfile --tag elastic/go-elasticsearch .
#
# $ docker run -it --network elasticsearch --volume $PWD/tmp:/tmp:rw --rm elastic/go-elasticsearch gotestsum --format=short-verbose --junitfile=/tmp/integration-junit.xml -- --cover --coverprofile=/tmp/integration-coverage.out --tags='integration' -v ./...
#

ARG  VERSION=1-alpine
FROM golang:${VERSION}

RUN apk add --no-cache --quiet make curl git jq unzip tree && \
    go get -u golang.org/x/lint/golint && \
    curl -sSL --retry 3 --retry-connrefused https://github.com/gotestyourself/gotestsum/releases/download/v0.3.2/gotestsum_0.3.2_linux_amd64.tar.gz | tar -xz -C /usr/local/bin gotestsum

VOLUME ["/tmp"]

ENV CGO_ENABLED=0
ENV TERM xterm-256color

WORKDIR /go-elasticsearch
COPY . .

RUN go mod download && go mod vendor && \
    cd internal/cmd/generate && go mod download && go mod vendor
