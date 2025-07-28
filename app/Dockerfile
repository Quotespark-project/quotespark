# Build stage (use a glibc-based image, e.g., debian or ubuntu)
FROM golang:1.23.2 AS builder
WORKDIR /app
COPY . .
# Statically compile
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o quotespark .


# Run stage
# Use a minimal but functional base image
FROM debian:bullseye-slim
WORKDIR /app
COPY --from=builder /app/quotespark .
EXPOSE 8080
CMD ["./quotespark"]