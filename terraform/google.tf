provider "google" {
  project = "patrolbot-485721"
  region  = "us-central1"
}

resource "google_project_service" "kms" {
  service = "cloudkms.googleapis.com"
}

resource "google_kms_key_ring" "sops" {
  name       = "sops-keyring"
  location   = "global"
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
