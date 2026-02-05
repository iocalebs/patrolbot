provider "google" {
  project = local.google_project_id
  region  = local.google_region
}

resource "google_project_service" "artifact_registry" {
  service = "artifactregistry.googleapis.com"
}

resource "google_project_service" "iam" {
  service = "iam.googleapis.com"
}

resource "google_project_service" "kms" {
  service = "cloudkms.googleapis.com"
}

resource "google_project_service" "cloud_run" {
  service = "run.googleapis.com"
}

resource "google_project_service" "secretmanager" {
  service = "secretmanager.googleapis.com"
}

resource "google_artifact_registry_repository" "patrolbot" {
  repository_id = "patrolbot"
  format        = "DOCKER"

  docker_config {
    immutable_tags = true
  }

  depends_on = [
    google_project_service.artifact_registry
  ]
}

resource "google_artifact_registry_repository_iam_member" "ci" {
  repository = google_artifact_registry_repository.patrolbot.repository_id
  role       = "roles/artifactregistry.writer"
  member     = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.ci.name}/attribute.repository/${github_repository.patrolbot.full_name}"
}

resource "google_iam_workload_identity_pool" "ci" {
  workload_identity_pool_id = "ci-pool"
  display_name              = "CI/CD Pool"
  description               = "Workload identity pool for CI/CD runners"
  depends_on = [
    google_project_service.iam
  ]
}

resource "google_iam_workload_identity_pool_provider" "github" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.ci.workload_identity_pool_id
  workload_identity_pool_provider_id = "github"
  display_name                       = "GitHub Actions"
  description                        = "OIDC provider for GitHub Actions"

  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }

  attribute_mapping = {
    "google.subject"       = "assertion.sub"
    "attribute.repository" = "assertion.repository"
  }

  attribute_condition = "assertion.repository=='${github_repository.patrolbot.full_name}'"
}

resource "google_kms_key_ring" "sops" {
  name     = "sops-keyring"
  location = "global"
  depends_on = [
    google_project_service.kms
  ]
}

resource "google_kms_crypto_key" "sops" {
  name            = "sops-key"
  key_ring        = google_kms_key_ring.sops.id
  rotation_period = "7776000s" # 90 days
  purpose         = "ENCRYPT_DECRYPT"

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_kms_crypto_key_iam_member" "github_sops" {
  crypto_key_id = google_kms_crypto_key.sops.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.ci.name}/attribute.repository/${github_repository.patrolbot.full_name}"
}

resource "google_project_iam_member" "ci" {
  project = local.google_project_id
  role    = "roles/run.admin"
  member  = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.ci.name}/attribute.repository/${github_repository.patrolbot.full_name}"
}

resource "google_secret_manager_secret" "discord_token" {
  secret_id = "discord-token"
  replication {
    auto {}
  }
  depends_on = [google_project_service.secretmanager]
}

resource "google_secret_manager_secret" "botpassword_zwen" {
  secret_id = "botpassword-zwen"
  replication {
    auto {}
  }
  depends_on = [google_project_service.secretmanager]
}

resource "google_secret_manager_secret_iam_member" "patrolbot" {
  for_each = toset([
    google_secret_manager_secret.botpassword_zwen.id,
    google_secret_manager_secret.discord_token.id,
  ])

  secret_id = each.value
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.patrolbot.email}"
}

resource "google_service_account" "patrolbot" {
  account_id = "patrolbot"
  depends_on = [google_project_service.iam]
}

resource "google_service_account_iam_member" "ci" {
  service_account_id = google_service_account.patrolbot.name
  role               = "roles/iam.serviceAccountUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.ci.name}/attribute.repository/${github_repository.patrolbot.full_name}"
}
