provider "google" {
  project = local.gcp_project_id
  region  = "us-central1"
}

resource "google_project_service" "iam" {
  service = "iam.googleapis.com"
}

resource "google_project_service" "kms" {
  service = "cloudkms.googleapis.com"
}

resource "google_project_service" "secretmanager" {
  service = "secretmanager.googleapis.com"
}

resource "google_iam_workload_identity_pool" "ci" {
  workload_identity_pool_id = "ci-pool"
  display_name              = "CI/CD Pool"
  description               = "Workload identity pool for CI/CD runners"

  depends_on = [google_project_service.iam]
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

  depends_on = [google_project_service.kms]
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
