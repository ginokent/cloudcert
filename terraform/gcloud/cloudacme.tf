terraform {
  required_providers {
    google      = ">= 4.19.0"
    google-beta = ">= 4.19.0"
  }
}

variable "container_image" {
  type = string
  validation {
    condition     = length(var.container_image) > 0 && can(regex("[[:graph:]]+(:[[:graph:]]+)?", var.container_image))
    error_message = "The container_image value must be a valid container image like 'NAME[:TAG|@DIGEST]'."
  }
}

variable "region" {
  type    = string
  default = "asia-northeast1"
}

data "google_project" "project" {}

resource "google_service_account" "cloudacme" {
  project      = data.google_project.project.project_id
  account_id   = "cloudacme"
  display_name = "cloudacme service account"
  description  = "managed by Terraform."
}

resource "google_project_iam_custom_role" "cloudacme" {
  project     = data.google_project.project.project_id
  role_id     = "custom.cloudacme"
  title       = "Role for cloudacme"
  description = "The minimum set of permissions required by cloudacme."
  stage       = "GA"
  permissions = [
    # for OpenTelemetry Tracing
    "cloudtrace.traces.patch",
    # for Lego Let's Encrypt DNS-01 Challenge
    "dns.changes.create",
    "dns.changes.get",
    "dns.managedZones.list",
    "dns.resourceRecordSets.create",
    "dns.resourceRecordSets.delete",
    "dns.resourceRecordSets.list",
    # for Store private key and certificate
    "secretmanager.secrets.get",
    "secretmanager.secrets.update",
    "secretmanager.versions.access",
    "secretmanager.versions.add",
    "secretmanager.versions.get",
  ]
}

resource "google_project_iam_binding" "cloudacme" {
  project = data.google_project.project.project_id
  role    = google_project_iam_custom_role.cloudacme.id
  members = [
    "serviceAccount:${google_service_account.cloudacme.email}",
  ]
}

resource "google_cloud_run_service" "cloudacme" {
  name                       = "cloudacme"
  project                    = data.google_project.project.project_id
  autogenerate_revision_name = true
  location                   = var.region
  metadata {
    namespace = data.google_project.project.project_id
    annotations = {
      "client.knative.dev/user-image"     = var.container_image
      "run.googleapis.com/client-name"    = "terraform"
      "run.googleapis.com/client-version" = "terraform"
      "run.googleapis.com/ingress"        = "all"
    }
  }
  template {
    metadata {
      annotations = {
        "autoscaling.knative.dev/maxScale"  = 1
        "client.knative.dev/user-image"     = var.container_image
        "run.googleapis.com/client-name"    = "terraform"
        "run.googleapis.com/client-version" = "terraform"
        # "run.googleapis.com/cpu-throttling" = true
        # "run.googleapis.com/execution-environment" = "gen1"
      }
    }
    spec {
      containers {
        image = var.container_image
        env {
          name  = "GOOGLE_CLOUD_PROJECT"
          value = data.google_project.project.project_id
        }
        env {
          name  = "SPAN_EXPORTER"
          value = "gcloud"
        }
        resources {
          limits = {
            cpu    = "1000m"
            memory = "512Mi"
          }
          requests = {}
        }
      }
      service_account_name = google_service_account.cloudacme.email
    }
  }
  traffic {
    percent         = 100
    latest_revision = true
  }
  timeouts {
    create = "10m"
    update = "10m"
    delete = "20m"
  }
  lifecycle {
    ignore_changes = [
      # metadata[0].annotations["client.knative.dev/user-image"],
      # metadata[0].annotations["run.googleapis.com/client-name"],
      # metadata[0].annotations["run.googleapis.com/client-version"],
      # template[0].metadata[0].annotations["run.googleapis.com/client-name"],
      # template[0].metadata[0].annotations["run.googleapis.com/client-version"],
      # template[0].metadata[0].annotations["run.googleapis.com/execution-environment"],
      # template[0].metadata[0].annotations["run.googleapis.com/sandbox"],
      # metadata,
      # template,
      # traffic,
    ]
  }
}

resource "google_cloud_run_service_iam_member" "cloudacme" {
  location = google_cloud_run_service.cloudacme.location
  project  = google_cloud_run_service.cloudacme.project
  service  = google_cloud_run_service.cloudacme.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.cloudacme.email}"
}
