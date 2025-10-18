# Cloud Region/Zone Generators

This directory contains generators for creating cloud provider region and zone mapping files.

## Overview

The generators fetch region/zone information from cloud provider CLIs and generate Go files with:
- Region/zone mappings to integer IDs
- Top regions for global coverage (prioritized for lower IDs)
- Display names and metadata as comments

## Generators

### GCP Generator (`gcpgen.go`)
Generates `gcpzones.go` using `gcloud compute zones list`.

**Usage:**
```bash
# Use defaults (8 global regions)
go run gcpgen.go

# Custom top zones (must be 2, 4, or 8)
go run gcpgen.go -top-zones "us-central1:a,europe-west1:b,asia-east1:a,australia-southeast1:a"
```

### AWS Generator (`awsgen.go`)
Generates `awsregions.go` using AWS SDK.

**Usage:**
```bash
# Use defaults (8 global regions)
go run awsgen.go

# Custom top regions (must be 2, 4, or 8)
go run awsgen.go -top-regions "us-east-1,eu-west-1,ap-southeast-1,sa-east-1"
```

### Azure Generator (`azuregen.go`)
Generates `azureregions.go` using `az account list-locations`.

**Usage:**
```bash
# Use defaults (8 global regions)
go run azuregen.go

# Custom top regions (must be 2, 4, or 8)
go run azuregen.go -top-regions "eastus,westeurope,southeastasia,australiaeast"
```

## Generate All

Use `generate_all.sh` to generate all cloud provider files at once.

**Usage:**
```bash
# Generate all with defaults
./generate_all.sh

# Generate with custom regions for specific providers
./generate_all.sh --azure-top-regions "eastus,westeurope"

# Generate with custom regions for all providers
./generate_all.sh \
  --gcp-top-zones "us-central1:a,europe-west1:b,asia-east1:a,australia-southeast1:a" \
  --aws-top-regions "us-east-1,eu-west-1,ap-southeast-1,sa-east-1" \
  --azure-top-regions "eastus,westeurope,southeastasia,australiaeast"

# Show help
./generate_all.sh --help
```

## Requirements

- **GCP:** `gcloud` CLI configured and authenticated
- **AWS:** AWS credentials configured (via `~/.aws/credentials` or environment variables)
- **Azure:** `az` CLI configured and authenticated

## Top Regions

Top regions are prioritized to get the first IDs, ensuring global coverage even when using limited bits (e.g., 3 bits = 8 regions) for cluster IDs.

Default top regions provide global geographic distribution:
- **GCP:** us-central1, europe-north1, asia-northeast1, australia-southeast2, southamerica-east1, africa-south1, me-west1, asia-south2
- **AWS:** us-west-1, eu-west-2, ap-southeast-2, sa-east-1, af-south-1, me-central-1, ap-east-1, ca-central-1
- **Azure:** eastus, westeurope, southeastasia, australiaeast, brazilsouth, southafricanorth, uaenorth, centralindia
