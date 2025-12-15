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

# Download dependencies
RUN go mod download

# Copy server source code
COPY server/ ./server/

# Build the server binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /bouncebot-server ./server

# Runtime stage
FROM alpine:latest

WORKDIR /

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /bouncebot-server /bouncebot-server

# Expose the server port
EXPOSE 8080

# Run the server
ENTRYPOINT ["/bouncebot-server"]
