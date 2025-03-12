ARG BUILD_PLATFORM="linux/amd64"
ARG TARGET_ARCH=amd64
ARG TARGET_OS=linux
ARG VERSION=dev

FROM --platform=${BUILD_PLATFORM} golang:1.23-alpine AS builder

RUN adduser -u 1001 -D valkyrie

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ENV CGO_ENABLED=0
ENV GOARCH=${TARGET_ARCH}
ENV GOOS=${TARGET_OS}

RUN GOARCH=${GOARCH} GOOS=${GOOS} go build -ldflags="-w -s -X main.appVersion=${VERSION}" -o valkyrie

FROM alpine:3.21

RUN apk add --no-cache curl

WORKDIR /app
# Add CA certs which are missing in scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/valkyrie /usr/bin/
COPY --from=builder /etc/passwd /etc/passwd

USER 1001

ENTRYPOINT ["/usr/bin/valkyrie"]
