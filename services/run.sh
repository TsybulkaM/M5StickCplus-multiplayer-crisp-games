#!/bin/bash

# Configuration
export DATABASE_URL="postgres://user:password@localhost:5432/crisp_db?sslmode=disable"
export MQTT_BROKER="tcp://localhost:1883"
export STORAGE_TYPE="local"
export LOCAL_STORAGE_PATH="./firmware_storage"
export ADMIN_API_TOKEN="dev-token-12345"

# Create storage directory if it doesn't exist
mkdir -p ./firmware_storage

# Build services
echo "Building services..."
go build -o engine-app ./cmd/engine
go build -o portal-app ./cmd/portal

# Run database migrations
echo "Running database migrations..."
./engine-app migrate

# Run engine in background
echo "Starting engine on port 8081..."
PORT=8081 \
DATABASE_URL="$DATABASE_URL" \
MQTT_BROKER="$MQTT_BROKER" \
STORAGE_TYPE="$STORAGE_TYPE" \
LOCAL_STORAGE_PATH="$LOCAL_STORAGE_PATH" \
ADMIN_API_TOKEN="$ADMIN_API_TOKEN" \
./engine-app &
ENGINE_PID=$!

# Run portal in background
echo "Starting portal on port 8080..."
PORT=8080 \
DATABASE_URL="$DATABASE_URL" \
MQTT_BROKER="$MQTT_BROKER" \
./portal-app &
PORTAL_PID=$!

echo ""
echo "=========================================="
echo "✓ Services started!"
echo "=========================================="
echo ""
echo "Portal: http://localhost:8080"
echo ""
echo "Engine API: http://localhost:8081"
echo "  - Health: http://localhost:8081/health"
echo "  - FOTA Check: http://localhost:8081/api/fota/check"
echo "  - FOTA Upload: http://localhost:8081/api/fota/upload (requires token)"
echo ""
echo "Configuration:"
echo "  - Database: localhost:5432/crisp_db"
echo "  - MQTT: localhost:1883"
echo "  - Storage: ./firmware_storage (local)"
echo "  - Admin Token: $ADMIN_API_TOKEN"
echo ""
echo "Engine PID: $ENGINE_PID"
echo "Portal PID: $PORTAL_PID"
echo ""
echo "Press Ctrl+C to stop services"
echo "=========================================="
echo ""

# Trap Ctrl+C to kill both processes
trap "echo ''; echo 'Stopping services...'; kill $ENGINE_PID $PORTAL_PID 2>/dev/null; exit" INT

# Wait for both processes
wait
