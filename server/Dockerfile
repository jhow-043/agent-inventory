FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY shared/ shared/
COPY server/ server/

WORKDIR /build/server

ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/server ./cmd/api

FROM alpine:3.20

RUN apk --no-cache add ca-certificates wget tzdata

COPY --from=builder /app/server /app/server

WORKDIR /app

EXPOSE 8080

CMD ["./server"]
