#!/usr/bin/env bash
set -euo pipefail

# Make sure we're in the right directory
cd "$(dirname "$0")"

# Parse command line arguments
GCP_TOP_ZONES=""
AWS_TOP_REGIONS=""
AZURE_TOP_REGIONS=""

usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Generate cloud region/zone files for GCP, AWS, and Azure."
    echo ""
    echo "Options:"
    echo "  --gcp-top-zones ZONES       GCP top zones in format 'region1:zone1,region2:zone2'"
    echo "                              Example: 'us-central1:a,europe-west1:b'"
    echo "  --aws-top-regions REGIONS   AWS top regions in comma-separated format"
    echo "                              Example: 'us-east-1,eu-west-1,ap-southeast-1'"
    echo "  --azure-top-regions REGIONS Azure top regions in comma-separated format"
    echo "                              Example: 'eastus,westeurope,southeastasia'"
    echo "  -h, --help                  Show this help message"
    echo ""
    echo "If no options are provided, default regions will be used for all clouds."
    exit 1
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --gcp-top-zones)
            GCP_TOP_ZONES="$2"
            shift 2
            ;;
        --aws-top-regions)
            AWS_TOP_REGIONS="$2"
            shift 2
            ;;
        --azure-top-regions)
            AZURE_TOP_REGIONS="$2"
            shift 2
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

# Generate GCP zones file
echo "Generating GCP zones file..."
if [ -z "$GCP_TOP_ZONES" ]; then
    echo "Using default hardcoded zones..."
    go run gcpgen.go
else
    echo "Using custom top zones: $GCP_TOP_ZONES"
    go run gcpgen.go -top-zones "$GCP_TOP_ZONES"
fi

# Generate AWS regions file
echo ""
echo "Generating AWS regions file..."
if [ -z "$AWS_TOP_REGIONS" ]; then
    echo "Using default hardcoded regions..."
    go run awsgen.go
else
    echo "Using custom top regions: $AWS_TOP_REGIONS"
    go run awsgen.go -top-regions "$AWS_TOP_REGIONS"
fi

# Generate Azure regions file
echo ""
echo "Generating Azure regions file..."
if [ -z "$AZURE_TOP_REGIONS" ]; then
    echo "Using default hardcoded regions..."
    go run azuregen.go
else
    echo "Using custom top regions: $AZURE_TOP_REGIONS"
    go run azuregen.go -top-regions "$AZURE_TOP_REGIONS"
fi

# Format the generated files
go fmt ../

echo ""
echo "All cloud zone/region files generated successfully!"
