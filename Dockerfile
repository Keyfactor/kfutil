# Build the kfutil binary
FROM golang:1.20 as builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY cmd/ cmd/
COPY pkg/ pkg/

# Build
# the GOARCH has a default value to allow the binary be built according to the host where the command
# was called. For example, if we build in a local env which has the Apple Silicon M1 chip,
# the docker BUILDPLATFORM arg will be linux/arm64, but for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o kfutil main.go

FROM alpine:latest
WORKDIR /
COPY --from=builder /workspace/kfutil /usr/local/bin/kfutil

# Spin forever so that the container doesn't exit and the user can exec into it
ENTRYPOINT ["tail", "-f", "/dev/null"]