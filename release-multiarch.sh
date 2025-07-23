#!/bin/bash

# Script to build and push multi-architecture Docker images for unix-user-exporter

set -e

# Check if version is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <version>"
  echo "Example: $0 v1.1.0"
  exit 1
fi

VERSION=$1
REPO="zerepl/unix-user-exporter"

echo "Setting up Docker buildx for multi-architecture builds..."

# Create a new builder instance if it doesn't exist
if ! docker buildx inspect multiarch > /dev/null 2>&1; then
  docker buildx create --name multiarch --use
else
  docker buildx use multiarch
fi

# Make sure the builder is running
docker buildx inspect --bootstrap

echo "Building and pushing multi-architecture images for unix-user-exporter $VERSION..."

# Build and push the multi-architecture images
docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 \
  --tag $REPO:$VERSION \
  --tag $REPO:latest \
  --push \
  .

echo "Done! Multi-architecture images pushed to Docker Hub:"
echo "- $REPO:$VERSION (linux/amd64, linux/arm64, linux/arm/v7)"
echo "- $REPO:latest (linux/amd64, linux/arm64, linux/arm/v7)"
