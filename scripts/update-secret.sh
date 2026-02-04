#!/usr/bin/env bash
set -euo pipefail

# Usage: ./update-secret.sh SECRET_NAME ENV_VAR_NAME

if [ $# -ne 2 ]; then
  echo "Usage: $0 <secret-name> <env-var-name>"
  exit 1
fi

SECRET_NAME="$1"
ENV_VAR_NAME="$2"

SECRET_VALUE="${!ENV_VAR_NAME:-}"

if [ -z "$SECRET_VALUE" ]; then
  echo "Error: Environment variable '$ENV_VAR_NAME' is empty or not set."
  exit 1
fi

read -p "Add new version to secret '$SECRET_NAME' from env variable '$ENV_VAR_NAME'? [y/N]: " confirm
if [[ "$confirm" =~ ^[Yy]$ ]]; then
  printf "%s" "$SECRET_VALUE" | gcloud secrets versions add "$SECRET_NAME" --data-file=-
  echo ""
fi
