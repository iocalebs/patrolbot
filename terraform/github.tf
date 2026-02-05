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

resource "github_actions_secret" "codecov_token" {
  repository      = github_repository.patrolbot.name
  secret_name     = "CODECOV_TOKEN"
  plaintext_value = var.codecov_token
}


resource "github_actions_variable" "google_artifact_registry_host" {
  repository    = github_repository.patrolbot.name
  variable_name = "GOOGLE_ARTIFACT_REGISTRY_HOST"
  value         = split("/", google_artifact_registry_repository.patrolbot.registry_uri)[0]
}

resource "github_actions_variable" "google_artifact_registry_repository_uri" {
  repository    = github_repository.patrolbot.name
  variable_name = "GOOGLE_ARTIFACT_REGISTRY_REPO_URI"
  value         = google_artifact_registry_repository.patrolbot.registry_uri
}

resource "github_actions_variable" "google_project" {
  repository    = github_repository.patrolbot.name
  variable_name = "GOOGLE_PROJECT"
  value         = local.google_project_id
}

resource "github_actions_variable" "google_region" {
  repository    = github_repository.patrolbot.name
  variable_name = "GOOGLE_REGION"
  value         = local.google_region
}

resource "github_actions_variable" "google_service_account" {
  repository    = github_repository.patrolbot.name
  variable_name = "GOOGLE_SERVICE_ACCOUNT"
  value         = google_service_account.patrolbot.email
}

resource "github_actions_variable" "google_workload_identity_provider" {
  repository    = github_repository.patrolbot.name
  variable_name = "GOOGLE_WORKLOAD_IDENTITY_PROVIDER"
  value         = google_iam_workload_identity_pool_provider.github.name
}
