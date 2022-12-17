ARG TARGETARCH=amd64

### Base builder image for native builds architecture
FROM golang:1.17-alpine AS builder-native-base
ENV CGO_ENABLED=0 GOOS=linux


### Intermediate builder image for x86-64 native builds
FROM builder-native-base AS builder-for-amd64
ENV GOARCH=amd64


### Intermediate builder image for AArch64 native builds
FROM builder-native-base AS builder-for-arm64v8
ENV GOARCH=arm64


### Final builder image where the build happens
# Possible build strategies:
# TARGETARCH=amd64
# TARGETARCH=arm64v8
ARG TARGETARCH=amd64
FROM builder-for-${TARGETARCH} AS builder

WORKDIR /app/build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -ldflags="-extldflags=-static -s -w" -o hub .

### The shipped image
ARG TARGETARCH=amd64
FROM ${TARGETARCH}/busybox:latest

ENV GIN_MODE=release

WORKDIR /app/data/
WORKDIR /app

# Copy binary and config files from /build to root folder of scratch container.
COPY --from=builder ["/app/build/hub", "."]

ENTRYPOINT ["/app/hub"]
