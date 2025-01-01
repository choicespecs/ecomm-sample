#!/bin/bash

# Automatically detect service directories relative to the script's location
BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_DIRS=(
  "$BASE_DIR/inventory-service"
  "$BASE_DIR/notification-service"
  "$BASE_DIR/order-service"
)

# Step 1: Start Podman Compose
echo "Starting Podman Compose..."
podman-compose up -d
if [ $? -ne 0 ]; then
  echo "Failed to start Podman Compose. Exiting."
  exit 1
fi

# Step 2: Run each Go service
for SERVICE_DIR in "${SERVICE_DIRS[@]}"; do
  echo "Starting service in directory: $SERVICE_DIR"
  (
    cd "$SERVICE_DIR" || {
      echo "Failed to change directory to $SERVICE_DIR. Skipping."
      continue
    }
    go run main.go &
    echo "Service in $SERVICE_DIR started with PID $!"
  )
done

# Step 3: Wait for all background services to start
echo "All services are starting in the background. Use 'ps' to check running processes."
echo "To stop all services, use 'kill' with their PIDs or terminate this script."
