# Dockerfile para compilar o agent Windows
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY shared/ shared/
COPY agent/ agent/

WORKDIR /build/agent

RUN CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o /app/agent.exe ./cmd/agent

FROM scratch
COPY --from=builder /app/agent.exe /agent.exe
