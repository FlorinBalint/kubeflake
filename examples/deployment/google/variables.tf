// Resolve region from zonal var.location (e.g., us-central1-a -> us-central1)
locals {
  gcp_region = join("-", slice(split("-", var.location), 0, 2))
}

# Optional var; empty means "derive from gcloud"
variable "project_id" {
  description = "GCP project ID (optional; auto-detected if empty)"
  type        = string
  default     = ""
}

variable "location" {
  description = "GKE cluster location; must be a ZONE like us-central1-a"
  type        = string
  default     = "us-central1-a"

  validation {
    condition     = length(split("-", var.location)) == 3
    error_message = "Use a ZONE (e.g., europe-west2-b). Regional values (e.g., europe-west2) are not allowed."
  }
}

variable "cluster_name" {
  description = "GKE cluster name"
  type        = string
  default     = "kubeflake"
}

variable "node_count" {
  description = "Number of nodes in the default node pool"
  type        = number
  default     = 3
}

variable "node_machine_type" {
  description = "GCE machine type for nodes"
  type        = string
  default     = "e2-standard-2"
}

variable "node_disk_size_gb" {
  description = "Boot disk size (GB) for nodes"
  type        = number
  default     = 75
}

variable "node_disk_type" {
  description = "Boot disk type for nodes: pd-standard | pd-balanced | pd-ssd"
  type        = string
  default     = "pd-ssd"
}

# Workload parameters (no hardcoded app names)
variable "namespace" {
  description = "Kubernetes namespace for the app"
  type        = string
  default     = "kubeflake"
}

variable "app_name" {
  description = "Application name (used for resources: StatefulSet, Services, HPA)"
  type        = string
  default     = "kubeflake"
}

variable "min_replicas" {
  description = "Minimum replicas for HPA/StatefulSet"
  type        = number
  default     = 2
}

variable "max_replicas" {
  description = "Maximum replicas for HPA"
  type        = number
  default     = 5
}

variable "container_args" {
  description = "Container args (leave empty to use per-service defaults)"
  type        = list(string)
  default     = []
}

# Container image configuration
variable "dockerhub_image" {
  description = "Docker Hub image name (e.g., 'florinbalint/kubeflake-keygen')"
  type        = string
  default     = "florinbalint/kubeflake-keygen"
}

variable "image_tag" {
  description = "Image tag to deploy"
  type        = string
  default     = "latest"
}

data "google_client_config" "current" {}

locals {
  // Prefer explicit var, then external gcloud (if defined), then provider client config
  actual_project = coalesce(
    var.project_id,
    try(data.external.gcloud_project.result.project, ""),
    try(data.google_client_config.current.project, "")
  )
}