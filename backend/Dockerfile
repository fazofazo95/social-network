# ── Build stage ──
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o server .

# ── Run stage ──
FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /build/server .
COPY --from=builder /build/pkg/db/migrations ./pkg/db/migrations

RUN mkdir -p pkg/db uploads

EXPOSE 8080

CMD ["./server"]
