#!/bin/bash
set -e

BASE_SCHEMA=$1
HEAD_SCHEMA=$2
CONFIG_FILE=$3

echo "Running Substrate API Contract Guard..."
echo "Base Schema: $BASE_SCHEMA"
echo "Head Schema: $HEAD_SCHEMA"

if [ -n "$CONFIG_FILE" ] && [ -f "$CONFIG_FILE" ]; then
  echo "Using Config: $CONFIG_FILE"
  substrate diff "$BASE_SCHEMA" "$HEAD_SCHEMA" --config "$CONFIG_FILE" --format text
else
  substrate diff "$BASE_SCHEMA" "$HEAD_SCHEMA" --format text
fi
