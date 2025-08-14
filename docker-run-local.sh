#!/bin/bash

# Local Docker run script for testing
# This script builds and runs the Docker container locally for testing

set -e

APP_NAME="planning-poker"
LOCAL_IMAGE="planning-poker:local"

echo "🧪 Building and running Planning Poker locally with Docker"

# Build the image locally
echo "🏗️  Building Docker image locally..."
docker build -t ${LOCAL_IMAGE} .

if [ $? -ne 0 ]; then
    echo "❌ Docker build failed!"
    exit 1
fi

# Stop and remove existing container if it exists
echo "🧹 Cleaning up existing container..."
docker stop ${APP_NAME} 2>/dev/null || true
docker rm ${APP_NAME} 2>/dev/null || true

# Run the container
echo "🚀 Starting container..."
docker run -d \
    -p 8080:8080 \
    -v planning-poker-db:/app/data \
    --name ${APP_NAME} \
    ${LOCAL_IMAGE}

if [ $? -ne 0 ]; then
    echo "❌ Failed to start container!"
    exit 1
fi

echo "✅ Container started successfully!"
echo ""
echo "🌐 Local application is running at: http://localhost:8080"
echo ""
echo "📋 Useful commands:"
echo "   docker logs ${APP_NAME}          # View logs"
echo "   docker stop ${APP_NAME}          # Stop container"
echo "   docker restart ${APP_NAME}       # Restart container"
echo "   docker exec -it ${APP_NAME} sh   # Shell into container"