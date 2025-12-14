#!/bin/bash

# Build services
echo "Building services..."
go build -o engine ./cmd/engine
go build -o portal ./cmd/portal

# Run engine in background
echo "Starting engine on port 8081..."
PORT=8081 ./engine &
ENGINE_PID=$!

# Run portal in background
echo "Starting portal on port 8080..."
PORT=8080 ./portal &
PORTAL_PID=$!

echo "Services started!"
echo "Engine PID: $ENGINE_PID"
echo "Portal PID: $PORTAL_PID"
echo "Portal available at http://localhost:8080"
echo ""
echo "Press Ctrl+C to stop services"

# Trap Ctrl+C to kill both processes
trap "echo 'Stopping services...'; kill $ENGINE_PID $PORTAL_PID; exit" INT

# Wait for both processes
wait
