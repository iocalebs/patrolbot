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
