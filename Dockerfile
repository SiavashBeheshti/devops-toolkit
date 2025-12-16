# Build stage
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /devops-toolkit .

# Final stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /devops-toolkit /usr/local/bin/devops-toolkit

# Switch to non-root user
USER appuser

ENTRYPOINT ["devops-toolkit"]
CMD ["--help"]

