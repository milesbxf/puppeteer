# Build the manager binary
FROM golang:1.10.3 as builder

# Copy in the go src
WORKDIR /go/src/github.com/milesbxf/puppeteer
COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor/ vendor/

# Build
RUN mkdir -p bin
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/manager github.com/milesbxf/puppeteer/cmd/manager

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/plugin_gitsource_job github.com/milesbxf/puppeteer/cmd/plugins/gitsource/job

# Copy the controller-manager into a thin image
FROM debian:buster-slim

WORKDIR /bin

COPY --from=builder /go/src/github.com/milesbxf/puppeteer/bin .
