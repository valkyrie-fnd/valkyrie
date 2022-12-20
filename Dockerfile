FROM golang:1.19-alpine as builder
ARG VERSION
RUN apk --no-cache add ca-certificates git
ENV CGO_ENABLED=0
WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download
COPY . .
RUN GOOS=linux go build -ldflags="-w -s -X main.appVersion=${VERSION}" .

FROM scratch
WORKDIR /app
# Add CA certs which are missing in scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/valkyrie /usr/bin/
ENTRYPOINT ["valkyrie"]
