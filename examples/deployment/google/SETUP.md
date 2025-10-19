# GCP Terraform Setup Guide

## Quick Start

### 1. Configure GCP Authentication

Terraform needs to authenticate with GCP. Choose one of these methods:

#### Option A: Use gcloud CLI (Recommended for local development)

```bash
# Login to GCP
gcloud auth application-default login

# Set your default project
gcloud config set project YOUR_PROJECT_ID

# Verify
gcloud config get-value project
```

#### Option B: Use Service Account (For CI/CD)

```bash
# Create service account
gcloud iam service-accounts create terraform \
  --display-name "Terraform Service Account"

# Grant permissions
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
  --member="serviceAccount:terraform@YOUR_PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/container.admin"

gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
  --member="serviceAccount:terraform@YOUR_PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/compute.admin"

# Download key
gcloud iam service-accounts keys create terraform-key.json \
  --iam-account=terraform@YOUR_PROJECT_ID.iam.gserviceaccount.com

# Set environment variable
export GOOGLE_APPLICATION_CREDENTIALS="$(pwd)/terraform-key.json"
```

### 2. Create terraform.tfvars

The easiest way to configure Terraform is with a `terraform.tfvars` file:

```bash
cd examples/deployment/google

# Create terraform.tfvars
cat > terraform.tfvars <<EOF
# GCP Configuration
project_id   = "your-gcp-project-id"
location     = "us-central1-a"
cluster_name = "kubeflake"

# Docker Hub Configuration
dockerhub_image = "florinbalint/kubeflake-keygen"
image_tag       = "latest"

# Node Configuration
node_count        = 3
node_machine_type = "e2-standard-2"
node_disk_size_gb = 75
node_disk_type    = "pd-ssd"

# Application Configuration
namespace    = "kubeflake"
app_name     = "kubeflake"
min_replicas = 2
max_replicas = 5
EOF
```

### 3. Initialize and Deploy

```bash
# Initialize Terraform
terraform init

# See what will be created
terraform plan

# Create infrastructure
terraform apply
```

## Configuration Methods

Terraform will detect your GCP project in this order:

1. **terraform.tfvars** - `project_id = "your-project"`
2. **Environment variable** - `export TF_VAR_project_id=your-project`
3. **gcloud config** - Automatically detected from `gcloud config get-value project`

### Method 1: Using terraform.tfvars (Recommended)

```hcl
# terraform.tfvars
project_id = "my-gcp-project"
location   = "us-central1-a"
```

### Method 2: Using Environment Variables

```bash
export TF_VAR_project_id="my-gcp-project"
export TF_VAR_location="us-central1-a"

terraform plan
```

### Method 3: Using Command Line

```bash
terraform plan -var="project_id=my-gcp-project" -var="location=us-central1-a"
```

### Method 4: Let gcloud Auto-detect (If configured)

```bash
# Set default project
gcloud config set project my-gcp-project

# Terraform will auto-detect via data.google_client_config
terraform plan
```

## Required GCP APIs

Enable these APIs before running Terraform:

```bash
# Enable required APIs
gcloud services enable container.googleapis.com
gcloud services enable compute.googleapis.com
gcloud services enable iam.googleapis.com

# Verify
gcloud services list --enabled
```

## Permissions Required

Your GCP account (or service account) needs these roles:

- `roles/container.admin` - For GKE cluster management
- `roles/compute.admin` - For compute resources
- `roles/iam.serviceAccountAdmin` - For creating service accounts

Grant them:

```bash
# For your user
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
  --member="user:YOUR_EMAIL@example.com" \
  --role="roles/container.admin"

# Or for service account
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
  --member="serviceAccount:SA_NAME@PROJECT.iam.gserviceaccount.com" \
  --role="roles/container.admin"
```

## Troubleshooting

### Error: "project_id is required"

**Cause:** No project specified in any configuration method

**Fix:** Create `terraform.tfvars` with `project_id = "your-project"`

### Error: "API not enabled"

**Cause:** Required GCP APIs not enabled

**Fix:**
```bash
gcloud services enable container.googleapis.com compute.googleapis.com
```

### Error: "Permission denied"

**Cause:** Your account lacks required permissions

**Fix:**
```bash
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
  --member="user:YOUR_EMAIL" \
  --role="roles/container.admin"
```

### Error: "default credentials not found"

**Cause:** Not authenticated with GCP

**Fix:**
```bash
gcloud auth application-default login
```

## Complete Example

```bash
# 1. Authenticate
gcloud auth application-default login

# 2. Set project
gcloud config set project my-gcp-project

# 3. Enable APIs
gcloud services enable container.googleapis.com compute.googleapis.com

# 4. Create config
cd examples/deployment/google
cat > terraform.tfvars <<EOF
project_id      = "my-gcp-project"
location        = "us-central1-a"
dockerhub_image = "florinbalint/kubeflake-keygen"
image_tag       = "latest"
EOF

# 5. Deploy
terraform init
terraform plan
terraform apply
```

## Testing the Service

After deployment, test the keygen endpoint:

### Port-forward to your local machine

```bash
kubectl port-forward -n kubeflake svc/kubeflake-keygen-svc 8083:8083
```

Then in another terminal:

```bash
# Test health endpoint
curl http://localhost:8083/health

# Generate a key
curl http://localhost:8083/generate/v1
```

### Test from inside the cluster

```bash
kubectl run curl-test --image=curlimages/curl:latest --rm -i --restart=Never -n kubeflake -- \
  curl -s http://kubeflake-keygen-svc:8083/generate/v1
```

## What Gets Created

Terraform will create:

- GKE cluster (1 node pool)
- Kubernetes namespace
- Service account with Workload Identity
- StatefulSet (keygen server)
- Headless service
- Horizontal Pod Autoscaler

## Cleanup

To destroy everything:

```bash
terraform destroy
```

To destroy just the application:

```bash
terraform destroy -target=kubernetes_stateful_set.keygen
```

## See Also

- [Terraform Variables](terraform.tfvars.dockerhub.example)
- [GCP Documentation](https://cloud.google.com/docs)
