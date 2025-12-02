# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod ./
COPY pkg/ ./pkg/
COPY service-1-user/go.mod service-1-user/go.sum ./service-1-user/

# Download dependencies
WORKDIR /build/service-1-user
RUN go mod download

# Copy source code
COPY service-1-user/ ./

# Tidy dependencies
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o user-service ./cmd/server/

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/service-1-user/user-service .

# Copy migrations if needed
COPY --from=builder /build/service-1-user/migrations ./migrations

# Expose gRPC port
EXPOSE 50051

# Run the application
CMD ["./user-service"]
