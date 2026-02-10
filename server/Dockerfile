# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git (needed for some Go dependencies)
RUN apk add --no-cache git

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./

# Copy local packages that are referenced in go.mod replace directives
COPY model/ ./model/
COPY proto/ ./proto/
COPY solver/ ./solver/

# Download dependencies
RUN go mod download

# Copy server source code
COPY server/ ./server/

# Version is passed as build arg (defaults to dev)
ARG VERSION=dev

# Build the server binary with version embedded
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.Version=${VERSION}" -o /bouncebot-server ./server

# Runtime stage
FROM alpine:latest

WORKDIR /

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /bouncebot-server /bouncebot-server

# Environment variables for configuration
# PORT: Server port (default: 8080)
# DATA_DIR: Base directory for all data files (default: data)
# ALLOWED_ORIGINS: Comma-separated allowed origins (default: localhost)
# AUTO_SAVE_INTERVAL: Auto-save interval in seconds (default: 30)
# CLEANUP_INTERVAL: Cleanup interval in seconds (default: 3600)
# ROOM_MAX_AGE: Room max age in seconds (default: 86400)
# DISCONNECT_GRACE_PERIOD: Player disconnect grace period in seconds (default: 30)
# SOLO_DISCONNECT_GRACE_PERIOD: Solo mode disconnect grace period in seconds (default: 1800)
# ENABLE_DAILY_CHALLENGE: Enable daily challenge feature (default: false)

# Expose the server port
EXPOSE 8080

# Run the server
ENTRYPOINT ["/bouncebot-server"]
