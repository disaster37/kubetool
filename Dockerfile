# Build the kubetool binary
FROM golang:1.26 AS builder

ARG version
ARG commit

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY . ./

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags "-s -w -X main.version=${version} -X main.commit=${commit}" -o /workspace/kubetool-bin .

# Use alpine to have shell support
FROM alpine:3.22
WORKDIR /
COPY --from=builder /workspace/kubetool-bin /usr/local/bin/kubetool
RUN \
    apk --update add bash &&\
    rm -rf /var/cache/apk/*
USER 65532:65532
