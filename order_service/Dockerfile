# Dockerfile for a microservice (e.g., inventory_service)

# Stage 1: Build the Go binary
FROM docker.io/library/golang:1.20 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

# Stage 2: Create a runtime image
FROM docker.io/library/debian:bookworm-slim

WORKDIR /root/

# Install PostgreSQL
RUN apt-get update && apt-get install -y \
    postgresql \
    postgresql-contrib && \
    rm -rf /var/lib/apt/lists/*

# Copy the built binary
COPY --from=builder /app/main .

# Copy the initial PostgreSQL setup script
COPY init-db.sh /root/
RUN chmod +x /root/init-db.sh

# Expose the PostgreSQL port (optional if needed for external connections)
EXPOSE 5432

# Start PostgreSQL, wait for it to initialize, and then start the Go application
CMD service postgresql start && \
    /root/init-db.sh && \
    until pg_isready -h localhost -p 5432 -U postgres; do sleep 1; done && \
    ./main