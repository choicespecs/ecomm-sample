#!/bin/bash

# Automatically detect service directories relative to the script's location
BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_DIRS=(
  "$BASE_DIR/inventory_service"
  "$BASE_DIR/notification_service"
  "$BASE_DIR/order_service"
  "$BASE_DIR/api_gateway"
)

# Step 1: Stop podman compose
echo "Starting Podman Compose..."
podman-compose down
if [ $? -ne 0 ]; then
  echo "Failed to Podman Compose down. Exiting."
  exit 1
fi

# Step 2: Remove each service.log file
for SERVICE_DIR in "${SERVICE_DIRS[@]}"; do
  echo "Removing logs within: $SERVICE_DIR"
  (
    cd "$SERVICE_DIR" || {
      echo "Failed to change directory to $SERVICE_DIR. Skipping."
      continue
    }
    LOG_FILE="$SERVICE_DIR/service.log"
    echo "Deleting $LOG_FILE"
    rm -rf $LOG_FILE
    echo "$LOG_FILE deleted"
  )
done

# Find the process running on port 8080
PID=$(lsof -t -i:8080)

if [ -n "$PID" ]; then
  echo "Stopping process on port 8080 with PID $PID"
  kill -9 $PID
  echo "Process stopped."
else
  echo "No process running on port 8080."
fi