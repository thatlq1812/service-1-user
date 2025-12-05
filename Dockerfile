# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build
# Copy go.mod files (context is the service folder)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code from current context into the build image
COPY . ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o user-service ./cmd/server/

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/user-service .

# Copy migrations if needed
COPY --from=builder /build/migrations ./migrations

# Expose gRPC port
EXPOSE 50051

# Run the application
CMD ["./user-service"]
