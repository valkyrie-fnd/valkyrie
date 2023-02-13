FROM --platform=$BUILDPLATFORM golang:1.19-alpine as builder
ARG VERSION
RUN adduser -u 1001 -D valkyrie && apk --no-cache add ca-certificates=20220614-r4
ENV CGO_ENABLED=0
WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download
COPY . .

ARG TARGETOS TARGETARCH
ENV GOOS $TARGETOS
ENV GOARCH $TARGETARCH

RUN GOOS=linux go build -ldflags="-w -s -X main.appVersion=${VERSION}" .

FROM scratch
WORKDIR /app
# Add CA certs which are missing in scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/valkyrie /usr/bin/
COPY --from=builder /etc/passwd /etc/passwd
USER 1001
ENTRYPOINT ["valkyrie"]
