# Stage 1: Build
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./
# (If we had go.sum, we would copy it too)
# COPY go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o qdrant-exporter .

# Stage 2: Final Image
FROM alpine:latest

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/qdrant-exporter .

# Expose port
EXPOSE 9090

# Command to run
ENTRYPOINT ["./qdrant-exporter"]
