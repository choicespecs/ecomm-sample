# Dockerfile for a microservice (e.g., inventory_service)

# Stage 1: Build the Go binary
FROM docker.io/library/golang:1.20 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

# Stage 2: Create a lightweight runtime image
FROM docker.io/library/debian:bullseye-slim

WORKDIR /root/
COPY --from=builder /app/main .

CMD ["./main"]