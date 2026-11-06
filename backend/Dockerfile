# --- Build stage: compile the Go binary with CGO (required by mattn/go-sqlite3) ---
FROM golang:1.23-alpine AS builder

# build-base provides gcc + musl headers needed for CGO.
RUN apk add --no-cache build-base

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -ldflags "-s -w" -o server .

# --- Runtime stage: minimal image with the binary, migrations and sqlite CLI ---
FROM alpine:3.20

# ca-certificates for TLS; sqlite for the `sqlite3 <db>` inspection in the audit.
RUN apk add --no-cache ca-certificates sqlite

WORKDIR /app
COPY --from=builder /app/server ./server
COPY --from=builder /app/pkg/db/migrations ./pkg/db/migrations

RUN mkdir -p /app/data /app/uploads

ENV PORT=8080
ENV DB_PATH=/app/data/social.db
ENV UPLOAD_DIR=/app/uploads
ENV MIGRATIONS_PATH=pkg/db/migrations/sqlite

EXPOSE 8080
CMD ["./server"]
