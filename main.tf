terraform {
  cloud {
    organization = "zeldawiki"
    workspaces {
      name = "patrolbot"
    }
  }
  required_providers {
    github = {
        source = "integrations/github"
        version = "~> 6.0"
    }
    google = {
      source = "hashicorp/google"
      version = "~> 7.0"
    }
  }
}

provider "github" {
  owner = "iocalebs"
}

resource "github_repository" "patrolbot" {
  name = "patrolbot"
  description = "Discord bot for MediaWiki patrollers"
  visibility = "public"
  has_issues = true
  has_discussions = true
}

variable "codecov_token" {
  type = string
  sensitive = true
}

variable "discord_token" {
  type = string
  sensitive = true
}

variable "zwen_password" {
  type = string
  sensitive = true
}

resource "github_actions_secret" "codecov_token" {
  repository = github_repository.patrolbot.name
  secret_name = "CODECOV_TOKEN"
  plaintext_value = var.codecov_token
}

resource "github_actions_secret" "discord_token" {
  repository = github_repository.patrolbot.name
  secret_name = "DISCORD_TOKEN"
  plaintext_value = var.discord_token
}

resource "github_actions_secret" "zwen_password" {
  repository = github_repository.patrolbot.name
  secret_name = "BOTPASSWORD_ZWEN"
  plaintext_value = var.zwen_password
}

provider "google" {
  project = "patrolbot-485721"
  region = "us-central1"
}

resource "google_project_service" "kms" {
  service = "cloudkms.googleapis.com"
}

resource "google_kms_key_ring" "sops" {
  name = "sops-keyring"
  location = "global"
  depends_on = [ google_project_service.kms ]
}

resource "google_kms_crypto_key" "sops" {
  name = "sops-key"
  key_ring = google_kms_key_ring.sops.id
  rotation_period = "7776000s" # 90 days
  purpose = "ENCRYPT_DECRYPT"

  lifecycle {
    prevent_destroy = true
  }
}
