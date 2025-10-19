#!/usr/bin/env bash
set -euo pipefail

# Build and push script for kubeflake keygen server Docker image
# Usage: ./build-and-push.sh [OPTIONS]
#
# Options:
#   --image NAME          Docker image name (default: kubeflake-keygen)
#   --tag TAG             Image tag (default: latest)
#   --registry REGISTRY   Docker registry/username (required for push)
#   --platform PLATFORMS  Target platforms (default: linux/amd64,linux/arm64)
#   --push                Push to Docker Hub (default: build only)
#   --no-cache            Build without cache
#   -h, --help            Show this help message

# Change to repository root (two levels up from this script)
cd "$(dirname "$0")/../.."

# Default values
IMAGE_NAME="kubeflake-keygen"
TAG="latest"
REGISTRY=""
PLATFORMS="linux/amd64,linux/arm64"
PUSH=false
NO_CACHE=""

usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Build and optionally push kubeflake keygen server Docker image."
    echo ""
    echo "Options:"
    echo "  --image NAME          Docker image name (default: $IMAGE_NAME)"
    echo "  --tag TAG             Image tag (default: $TAG)"
    echo "  --registry REGISTRY   Docker registry/username (e.g., 'florinbalint' for Docker Hub)"
    echo "  --platform PLATFORMS  Target platforms (default: $PLATFORMS)"
    echo "  --push                Push to registry after build"
    echo "  --no-cache            Build without using cache"
    echo "  -h, --help            Show this help message"
    echo ""
    echo "Examples:"
    echo "  # Build only (local)"
    echo "  $0"
    echo ""
    echo "  # Build and push to Docker Hub"
    echo "  $0 --registry florinbalint --push"
    echo ""
    echo "  # Build with custom tag"
    echo "  $0 --tag v1.0.0 --registry florinbalint --push"
    echo ""
    echo "  # Build for single platform (faster)"
    echo "  $0 --platform linux/amd64"
    exit 1
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --image)
            IMAGE_NAME="$2"
            shift 2
            ;;
        --tag)
            TAG="$2"
            shift 2
            ;;
        --registry)
            REGISTRY="$2"
            shift 2
            ;;
        --platform)
            PLATFORMS="$2"
            shift 2
            ;;
        --push)
            PUSH=true
            shift
            ;;
        --no-cache)
            NO_CACHE="--no-cache"
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

# Validate registry if pushing
if [ "$PUSH" = true ] && [ -z "$REGISTRY" ]; then
    echo "Error: --registry is required when using --push"
    echo "Example: $0 --registry florinbalint --push"
    exit 1
fi

# Build full image name
if [ -n "$REGISTRY" ]; then
    FULL_IMAGE="$REGISTRY/$IMAGE_NAME:$TAG"
else
    FULL_IMAGE="$IMAGE_NAME:$TAG"
fi

echo "========================================="
echo "Kubeflake Keygen Server - Docker Build"
echo "========================================="
echo "Image:     $FULL_IMAGE"
echo "Platforms: $PLATFORMS"
echo "Push:      $PUSH"
echo "Cache:     $([ -n "$NO_CACHE" ] && echo "disabled" || echo "enabled")"
echo "========================================="
echo ""

# Check if docker is available
if ! command -v docker &> /dev/null; then
    echo "Error: docker is not installed or not in PATH"
    exit 1
fi

# Check if buildx is available for multi-platform builds
if ! docker buildx version &> /dev/null; then
    echo "Warning: docker buildx not available, falling back to regular build"
    echo "Multi-platform builds will not be supported."

    # Regular docker build (single platform)
    echo "Building image with docker build..."
    docker build $NO_CACHE \
        -f examples/build/Dockerfile \
        -t "$FULL_IMAGE" \
        .

    if [ "$PUSH" = true ]; then
        echo ""
        echo "Pushing image to registry..."
        docker push "$FULL_IMAGE"
    fi
else
    # Create or use existing buildx builder
    BUILDER_NAME="kubeflake-builder"
    if ! docker buildx inspect "$BUILDER_NAME" &> /dev/null; then
        echo "Creating buildx builder: $BUILDER_NAME"
        docker buildx create --name "$BUILDER_NAME" --use
    else
        docker buildx use "$BUILDER_NAME"
    fi

    # Build with buildx (supports multi-platform)
    BUILD_ARGS=(
        "buildx"
        "build"
        "--platform" "$PLATFORMS"
        "-f" "examples/build/Dockerfile"
        "-t" "$FULL_IMAGE"
    )

    # Add --no-cache flag only if set
    if [ -n "$NO_CACHE" ]; then
        BUILD_ARGS+=("$NO_CACHE")
    fi

    # Add push or load flag
    if [ "$PUSH" = true ]; then
        BUILD_ARGS+=("--push")
    else
        # Note: --load only works with single platform builds
        # For multi-platform, images stay in the build cache
        if [[ "$PLATFORMS" == *","* ]]; then
            echo "Note: Multi-platform build without push - images will stay in build cache"
            echo "      To load locally, use --platform linux/amd64 (or your platform)"
        else
            BUILD_ARGS+=("--load")
        fi
    fi

    # Add context path
    BUILD_ARGS+=(".")

    echo "Building image with docker buildx..."
    docker "${BUILD_ARGS[@]}"
fi

echo ""
echo "========================================="
echo "Build completed successfully!"
echo "========================================="
echo "Image: $FULL_IMAGE"

if [ "$PUSH" = true ]; then
    echo "Status: Pushed to registry"
    echo ""
    echo "Pull with:"
    echo "  docker pull $FULL_IMAGE"
else
    echo "Status: Built locally (not pushed)"
    echo ""
    echo "Run with:"
    echo "  docker run -p 8083:8083 $FULL_IMAGE"
    echo ""
    echo "To push to registry:"
    echo "  $0 --registry YOUR_USERNAME --push"
fi
echo "========================================="
