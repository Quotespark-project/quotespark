# Build stage
FROM golang:1.23.2 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o quotespark .

# Run stage - use a standard Linux base image
FROM ubuntu:20.04
WORKDIR /app

# Install CA certificates for HTTPS connections
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/quotespark .
COPY --from=builder /app/templates ./templates
EXPOSE 8080
CMD ["./quotespark"]