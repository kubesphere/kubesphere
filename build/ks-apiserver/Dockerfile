# Build
FROM golang:1.24.3 AS build_context

RUN apt-get update && apt-get install -y rsync
WORKDIR /workspace
ADD . /workspace/

RUN make ks-apiserver

FROM alpine:3.21.3

COPY --from=build_context /workspace/_output/bin/ks-apiserver /usr/local/bin/

WORKDIR /app

EXPOSE 9090
CMD ["sh"]
