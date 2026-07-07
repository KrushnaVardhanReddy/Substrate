#!/bin/bash
set -e

BASE_SCHEMA=$1
HEAD_SCHEMA=$2
CONFIG_FILE=$3

echo "Running Substrate API Contract Guard..."

# Avoid "dubious ownership" errors in GitHub Actions
git config --global --add safe.directory /github/workspace

ACTUAL_BASE="$BASE_SCHEMA"

if [ -n "$GITHUB_BASE_REF" ]; then
  echo "Extracting base schema from origin/$GITHUB_BASE_REF..."
  # Fetch the base branch to ensure we have it locally
  git fetch origin "$GITHUB_BASE_REF" --depth=1 || true
  git show "origin/$GITHUB_BASE_REF:$BASE_SCHEMA" > /tmp/base_schema.yaml
  ACTUAL_BASE="/tmp/base_schema.yaml"
fi

echo "Base Schema: $ACTUAL_BASE (from $BASE_SCHEMA)"
echo "Head Schema: $HEAD_SCHEMA"

if [ -n "$CONFIG_FILE" ] && [ -f "$CONFIG_FILE" ]; then
  echo "Using Config: $CONFIG_FILE"
  substrate diff "$ACTUAL_BASE" "$HEAD_SCHEMA" --config "$CONFIG_FILE" --format text
else
  substrate diff "$ACTUAL_BASE" "$HEAD_SCHEMA" --format text
fi
