# build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o qdrant-exporter .

# final stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/qdrant-exporter .
EXPOSE 9999
CMD ["./qdrant-exporter"]
