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

echo ""
echo "Generating AWS regions file..."
# Check if top-regions flag is provided for AWS
if [ $# -lt 2 ]; then
    echo "Using default hardcoded regions..."
    # Run the AWS generator with default regions
    go run awsgen.go
else
    echo "Using custom top regions: $2"
    # Run the AWS generator with the provided top regions
    go run awsgen.go -top-regions "$2"
fi

# TODO: add generator for Azure

go fmt ../

echo ""
echo "All cloud zone/region files generated successfully!"
