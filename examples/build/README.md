# Kubeflake Keygen Server - Docker Build

This directory contains the Dockerfile and build scripts for the kubeflake keygen server example.

## Quick Start

### Build Locally

```bash
./build-and-push.sh
```

This builds the Docker image locally without pushing to a registry.

### Build and Push to Docker Hub

```bash
# First, login to Docker Hub
docker login

# Build and push
./build-and-push.sh --registry YOUR_DOCKERHUB_USERNAME --push
```

## Usage Examples

### Basic Build (Local Only)

```bash
./build-and-push.sh
```

Builds `kubeflake-keygen:latest` locally.

### Build with Custom Tag

```bash
./build-and-push.sh --tag v1.0.0
```

### Build and Push to Docker Hub

```bash
./build-and-push.sh \
  --registry florinbalint \
  --tag v1.0.0 \
  --push
```

### Build for Specific Platform

```bash
# AMD64 only (faster for local testing)
./build-and-push.sh --platform linux/amd64

# ARM64 only (for Apple Silicon, ARM servers)
./build-and-push.sh --platform linux/arm64

# Multi-platform (default)
./build-and-push.sh --platform linux/amd64,linux/arm64
```

### Build Without Cache

```bash
./build-and-push.sh --no-cache
```

## Running the Container

### Basic Run

```bash
docker run -p 8083:8083 kubeflake-keygen:latest
```

### Run with Custom Configuration

```bash
docker run -p 8083:8083 kubeflake-keygen:latest \
  --address :8083 \
  --bits.machine 6 \
  --bits.sequence 11 \
  --bits.cluster 7
```

### Run with Environment Variables (for testing)

```bash
# Test with GCP
docker run -p 8083:8083 \
  -e GCP_ZONE=us-central1-a \
  -e HOSTNAME=keygen-0 \
  kubeflake-keygen:latest

# Test with AWS
docker run -p 8083:8083 \
  -e AWS_REGION=us-east-1 \
  -e HOSTNAME=keygen-0 \
  kubeflake-keygen:latest

# Test with Azure
docker run -p 8083:8083 \
  -e AZURE_REGION=eastus \
  -e HOSTNAME=keygen-0 \
  kubeflake-keygen:latest
```

## Testing the Server

### Health Check

```bash
curl http://localhost:8083/health
```

### Generate a Key

```bash
curl http://localhost:8083/generate/v1
```

Example response:
```
2Kq3bN8pL5mR
```

## Dockerfile Details

### Multi-Stage Build

The Dockerfile uses a multi-stage build for optimal image size:

1. **Builder Stage** (`golang:1.23-alpine`):
   - Downloads dependencies
   - Builds static binary with optimizations

2. **Runtime Stage** (`alpine:latest`):
   - Minimal base image (~5MB)
   - Non-root user for security
   - CA certificates for metadata services
   - Health check configured

### Image Size

- Final image size: ~15-20MB
- Runs as non-root user (UID 1000)
- Includes health check endpoint

### Supported Platforms

- `linux/amd64` (Intel/AMD 64-bit)
- `linux/arm64` (ARM 64-bit, Apple Silicon, AWS Graviton)

## Build Script Options

```
--image NAME          Docker image name (default: kubeflake-keygen)
--tag TAG             Image tag (default: latest)
--registry REGISTRY   Docker registry/username (required for push)
--platform PLATFORMS  Target platforms (default: linux/amd64,linux/arm64)
--push                Push to Docker Hub (default: build only)
--no-cache            Build without cache
-h, --help            Show help message
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build and Push Docker Image

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Login to Docker Hub
        uses: docker/login-action@v2
        with:
          username: ${{ secrets.DOCKERHUB_USERNAME }}
          password: ${{ secrets.DOCKERHUB_TOKEN }}

      - name: Build and Push
        run: |
          cd examples/build
          ./build-and-push.sh \
            --registry ${{ secrets.DOCKERHUB_USERNAME }} \
            --tag ${GITHUB_REF#refs/tags/} \
            --push
```

## Troubleshooting

### Multi-Platform Build Without Push

When building for multiple platforms without `--push`, you'll see this note:

```
Note: Multi-platform build without push - images will stay in build cache
      To load locally, use --platform linux/amd64 (or your platform)
```

**Solution:** For local testing, build for a single platform:

```bash
./build-and-push.sh --platform linux/amd64
```

Or push to a registry to enable multi-platform:

```bash
./build-and-push.sh --registry florinbalint --push
```

### Docker Buildx Not Available

If you see a warning about buildx not being available:

```bash
# Install buildx plugin
docker buildx install

# Or use Docker Desktop which includes buildx
```

### Multi-Platform Build Issues

For multi-platform builds, you need:

1. Docker Buildx installed
2. QEMU for emulation:

```bash
docker run --privileged --rm tonistiigi/binfmt --install all
```

### Registry Authentication

Make sure you're logged in before pushing:

```bash
docker login
# Enter your Docker Hub username and password
```

### Build Fails with "requires 1 argument"

This was a bug in earlier versions. Make sure you have the latest version of the script where empty variables are handled correctly.

## Security Notes

- The container runs as a non-root user (UID 1000)
- Static binary with no dependencies
- Minimal attack surface (Alpine base)
- Health check included for orchestration
- CA certificates for secure metadata access
