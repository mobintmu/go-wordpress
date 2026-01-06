# Build and run stage
FROM golang:1.25.3-alpine

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code (use .dockerignore to exclude sensitive files)
# Copy only source code (not the entire directory)
COPY cmd/ ./cmd/
COPY docs/ ./docs/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY api/ ./api/


# Build the application and create non-root user in one layer
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/app/main.go && \
    adduser -D appuser

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run the application
CMD ["./server"]