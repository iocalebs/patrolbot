terraform {
  cloud {
    organization = "zeldawiki"
    workspaces {
      name = "patrolbot"
    }
  }
  required_providers {
    github = {
      source  = "integrations/github"
      version = "~> 6.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 7.0"
    }
  }
}
