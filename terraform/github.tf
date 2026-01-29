provider "github" {
  owner = "iocalebs"
}

resource "github_repository" "patrolbot" {
  name            = "patrolbot"
  description     = "Discord bot for MediaWiki patrollers"
  visibility      = "public"
  has_issues      = true
  has_discussions = true
}

resource "github_actions_variable" "google_project" {
  repository    = github_repository.patrolbot.name
  variable_name = "GOOGLE_PROJECT"
  value         = local.gcp_project_id
}

resource "github_actions_variable" "google_workload_identity_provider" {
  repository    = github_repository.patrolbot.name
  variable_name = "GOOGLE_WORKLOAD_IDENTITY_PROVIDER"
  value         = google_iam_workload_identity_pool_provider.github.name
}

resource "github_actions_secret" "codecov_token" {
  repository      = github_repository.patrolbot.name
  secret_name     = "CODECOV_TOKEN"
  plaintext_value = var.codecov_token
}

resource "github_actions_secret" "discord_token" {
  repository      = github_repository.patrolbot.name
  secret_name     = "DISCORD_TOKEN"
  plaintext_value = var.discord_token
}

resource "github_actions_secret" "zwen_password" {
  repository      = github_repository.patrolbot.name
  secret_name     = "BOTPASSWORD_ZWEN"
  plaintext_value = var.zwen_password
}
