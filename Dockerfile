FROM golang:1.10.3 as builder

# Copy in the go src
WORKDIR /go/src/kubesphere.io/kubesphere
COPY pkg/ pkg/
COPY cmd/ cmd/
COPY vendor/ vendor/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o ks-apiserver kubesphere.io/kubesphere/cmd/ks-apiserver


FROM alpine:3.6
WORKDIR /
COPY --from=builder /go/src/kubesphere.io/kubesphere/ks-apiserver .
COPY ./install/ingress-controller /etc/kubesphere/ingress-controller
COPY ./install/swagger-ui        /usr/lib/kubesphere/swagger-ui
CMD ["ks-apiserver"]
