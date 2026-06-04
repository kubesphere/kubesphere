FROM golang:1.18 as builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o ks-apiserver ./cmd/ks-apiserver

FROM alpine:3.16
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/ks-apiserver /usr/bin/
CMD ["ks-apiserver"]
