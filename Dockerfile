ARG TARGETARCH=amd64

### Base builder image for native builds architecture
FROM golang:1.17-alpine AS builder-native-base
ENV CGO_ENABLED=1 GOOS=linux
RUN apk add --no-cache g++ perl-utils


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

# Download Basenine executable, verify the sha1sum
ADD https://github.com/up9inc/basenine/releases/download/v0.8.3/basenine_linux_${GOARCH} ./basenine_linux_${GOARCH}
ADD https://github.com/up9inc/basenine/releases/download/v0.8.3/basenine_linux_${GOARCH}.sha256 ./basenine_linux_${GOARCH}.sha256

RUN shasum -a 256 -c basenine_linux_"${GOARCH}".sha256 && \
    chmod +x ./basenine_linux_"${GOARCH}" && \
    mv ./basenine_linux_"${GOARCH}" ./basenine

### The shipped image
ARG TARGETARCH=amd64
FROM ${TARGETARCH}/busybox:latest
# gin-gonic runs in debug mode without this
ENV GIN_MODE=release

WORKDIR /app/data/
WORKDIR /app

# Copy binary and config files from /build to root folder of scratch container.
COPY --from=builder ["/app/build/hub", "."]
COPY --from=builder ["/app/build/basenine", "/usr/local/bin/basenine"]

ENTRYPOINT ["/app/hub"]
