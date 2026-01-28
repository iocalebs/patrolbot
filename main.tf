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
