#!/usr/bin/env bash
# Quick build helper script for common scenarios

set -euo pipefail

cd "$(dirname "$0")"

SCRIPT="./build-and-push.sh"

case "${1:-}" in
    local)
        # Build locally for testing (single platform, fast)
        echo "Building locally for current platform..."
        $SCRIPT --platform linux/amd64
        ;;

    release)
        # Build and push release version
        if [ -z "${2:-}" ]; then
            echo "Error: Version required for release"
            echo "Usage: $0 release v1.0.0"
            exit 1
        fi
        if [ -z "${DOCKERHUB_USERNAME:-}" ]; then
            echo "Error: DOCKERHUB_USERNAME environment variable not set"
            echo "Example: export DOCKERHUB_USERNAME=florinbalint"
            exit 1
        fi

        TAG="$2"
        echo "Building and pushing release $TAG..."
        $SCRIPT \
            --registry "$DOCKERHUB_USERNAME" \
            --tag "$TAG" \
            --push

        # Also tag as latest
        echo ""
        echo "Also tagging as latest..."
        $SCRIPT \
            --registry "$DOCKERHUB_USERNAME" \
            --tag latest \
            --push
        ;;

    dev)
        # Build and push dev version
        if [ -z "${DOCKERHUB_USERNAME:-}" ]; then
            echo "Error: DOCKERHUB_USERNAME environment variable not set"
            echo "Example: export DOCKERHUB_USERNAME=florinbalint"
            exit 1
        fi

        TAG="dev-$(date +%Y%m%d-%H%M%S)"
        echo "Building and pushing dev version $TAG..."
        $SCRIPT \
            --registry "$DOCKERHUB_USERNAME" \
            --tag "$TAG" \
            --platform linux/amd64 \
            --push
        ;;

    test)
        # Build and run locally for testing
        echo "Building for local testing..."
        $SCRIPT --platform linux/amd64

        echo ""
        echo "Starting test container..."
        docker run --rm -d \
            -p 8083:8083 \
            -e GCP_ZONE=us-central1-a \
            -e HOSTNAME=keygen-0 \
            --name kubeflake-test \
            kubeflake-keygen:latest

        echo ""
        echo "Waiting for server to start..."
        sleep 2

        echo "Testing health endpoint..."
        curl -f http://localhost:8083/health && echo " ✓ Health check passed"

        echo ""
        echo "Generating test key..."
        KEY=$(curl -s http://localhost:8083/generate/v1)
        echo "Generated key: $KEY"

        echo ""
        echo "Stopping test container..."
        docker stop kubeflake-test

        echo ""
        echo "✓ All tests passed!"
        ;;

    *)
        echo "Quick build helper for kubeflake keygen server"
        echo ""
        echo "Usage: $0 <command> [options]"
        echo ""
        echo "Commands:"
        echo "  local              Build locally (fast, single platform)"
        echo "  dev                Build and push dev version with timestamp"
        echo "  release VERSION    Build and push release (e.g., v1.0.0)"
        echo "  test               Build, run, and test locally"
        echo ""
        echo "Environment Variables:"
        echo "  DOCKERHUB_USERNAME   Your Docker Hub username (required for dev/release)"
        echo ""
        echo "Examples:"
        echo "  $0 local"
        echo "  $0 test"
        echo "  export DOCKERHUB_USERNAME=florinbalint"
        echo "  $0 dev"
        echo "  $0 release v1.0.0"
        exit 1
        ;;
esac
