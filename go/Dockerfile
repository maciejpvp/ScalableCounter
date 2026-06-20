# --- Build Stage ---
FROM golang:1.25.11-alpine AS builder

# Install git and certificates
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy dependency manifests first to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the statically-linked binary with optimization flags
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o scalable-counter \
    cmd/api/main.go

# --- Final Stage ---
FROM alpine:3.20

# Install CA certificates for secure AWS API communication and tzdata for timezones
RUN apk add --no-cache ca-certificates tzdata

# Create a non-privileged user to run the application
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/scalable-counter .

# Change ownership of the application directory to the non-root user
RUN chown -R appuser:appgroup /app

# Run as the non-root user
USER appuser

# Set standard environment variables
ENV PORT=8080
ENV APP_ENV=production

# Expose the configured port
EXPOSE 8080

# Use exec form to ensure standard OS signals (SIGTERM, SIGINT) are received directly by the Go process
ENTRYPOINT ["./scalable-counter"]
