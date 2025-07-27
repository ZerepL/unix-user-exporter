#!/bin/bash

set -e

IMAGE_NAME="unix-user-exporter"
TAG="${1:-latest}"

echo "Building multi-architecture Docker image: ${IMAGE_NAME}:${TAG}"

# Check if buildx is available
if ! docker buildx version >/dev/null 2>&1; then
    echo "Docker buildx is not available. Building single architecture image..."
    docker build -t "${IMAGE_NAME}:${TAG}" .
    exit 0
fi

# Create a new builder instance if it doesn't exist
if ! docker buildx ls | grep -q multiarch-builder; then
    echo "Creating multiarch-builder..."
    docker buildx create --name multiarch-builder --use
fi

# Use the multiarch builder
docker buildx use multiarch-builder

# Build for multiple architectures
echo "Building for linux/amd64, linux/arm64, linux/arm/v7..."
docker buildx build \
    --platform linux/amd64,linux/arm64,linux/arm/v7 \
    --file Dockerfile.multiarch \
    --tag "${IMAGE_NAME}:${TAG}" \
    --load \
    .

echo "Multi-architecture build complete!"
echo "To test on different architectures:"
echo "  docker run --rm -p 32142:32142 -v /var/run/utmp:/var/run/utmp:ro ${IMAGE_NAME}:${TAG}"
