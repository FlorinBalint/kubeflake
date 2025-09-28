#!/usr/bin/env bash
set -euo pipefail

# Make sure we're in the right directory
cd "$(dirname "$0")"

echo "Generating GCP zones file..."
# Check if top-zones flag is provided
if [ $# -eq 0 ]; then
    echo "Using default hardcoded regions..."
    # Run the generator with default regions
    go run gcpgen.go
else
    echo "Using custom top zones: $1"
    # Run the generator with the provided top zones
    go run gcpgen.go -top-zones "$1"
fi
# TODO: add generators for AWS and Azure

go fmt ../

echo "Generation completed successfully!"
