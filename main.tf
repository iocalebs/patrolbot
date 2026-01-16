terraform {
  required_providers {
    github = {
        source = "integrations/github"
        version = "~> 6.0"
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

resource "github_actions_secret" "codecov_token" {
  repository = github_repository.patrolbot.name
  secret_name = "CODECOV_TOKEN"
  plaintext_value = var.codecov_token
}
