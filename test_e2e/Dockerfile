# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mqtt2ntfy .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/mqtt2ntfy .

# Copy config file
COPY test_e2e/config/mqtt2ntfy.yaml ./config.yaml

CMD ["./mqtt2ntfy", "--config", "config.yaml"]
