#!/bin/bash

# Docker deployment script for Planning Poker app
# This script builds and pushes the Docker image to Docker Hub

set -e

# Configuration
DOCKER_USERNAME="cougargriff"
APP_NAME="planning-poker"
TAG="latest"
PLATFORM="linux/amd64"
IMAGE_NAME="${DOCKER_USERNAME}/${APP_NAME}:${TAG}"

echo "🚀 Starting Docker deployment for Planning Poker app"
echo "Target: ${IMAGE_NAME}"
echo "Platform: ${PLATFORM}"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Error: Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if we're logged into Docker Hub
echo "🔐 Checking Docker Hub authentication..."
if ! docker info | grep -q "Username"; then
    echo "⚠️  Not logged into Docker Hub. Please log in:"
    docker login
fi

# Build the Docker image for the specified platform
echo "🏗️  Building Docker image..."
docker build --platform ${PLATFORM} -t ${IMAGE_NAME} .

if [ $? -ne 0 ]; then
    echo "❌ Docker build failed!"
    exit 1
fi

echo "✅ Docker image built successfully!"

# Get image size
IMAGE_SIZE=$(docker images ${IMAGE_NAME} --format "table {{.Size}}" | tail -n +2)
echo "📦 Image size: ${IMAGE_SIZE}"

# Push to Docker Hub
echo "🚀 Pushing image to Docker Hub..."
docker push ${IMAGE_NAME}

if [ $? -ne 0 ]; then
    echo "❌ Docker push failed!"
    exit 1
fi

echo "✅ Successfully pushed to Docker Hub!"
echo ""
echo "🎉 Deployment completed successfully!"
echo ""
echo "📋 To pull and run this image on your server:"
echo "   docker pull ${IMAGE_NAME}"
echo "   docker run -d -p 80:8080 --name planning-poker ${IMAGE_NAME}"
echo ""
echo "🔧 To run with persistent database:"
echo "   docker run -d -p 80:8080 -v planning-poker-db:/app/data --name planning-poker ${IMAGE_NAME}"
echo ""
echo "🌐 Access the application at: http://poker.griff.la"