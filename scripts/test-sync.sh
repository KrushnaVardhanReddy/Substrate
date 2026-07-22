#!/usr/bin/env bash
# test-sync.sh — Layer 2: API Sync + Dependency Graph Seeding
# Seeds the dependency graph for all 7 demo repos and verifies the graph.
# Usage: ./scripts/test-sync.sh [--verbose]

set -euo pipefail

API_URL="${API_URL:-http://localhost:8090}"
TOKEN="${INTERNAL_SERVICE_TOKEN:-local-dev-token}"
PAYLOAD_DIR="$(dirname "$0")/test-payloads"
ORG="local-demo-testing"
INSTALLATION_ID=99999
VERBOSE=false

for arg in "$@"; do
  case "$arg" in
    --verbose) VERBOSE=true ;;
  esac
done

PASS=0
FAIL=0
TOTAL=0

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_pass() { PASS=$((PASS + 1)); TOTAL=$((TOTAL + 1)); echo -e "  ${GREEN}✓ PASS${NC} $1"; }
log_fail() { FAIL=$((FAIL + 1)); TOTAL=$((TOTAL + 1)); echo -e "  ${RED}✗ FAIL${NC} $1: $2"; }
log_section() { echo -e "\n${YELLOW}━━━ $1 ━━━${NC}"; }

# Helper: call API with auth
api_post() {
  local path="$1"
  local data="$2"
  curl -s -X POST "${API_URL}${path}" \
    -H "Authorization: Bearer ${TOKEN}" \
    -H 'Content-Type: application/json' \
    -d "$data"
}

api_get() {
  local path="$1"
  curl -s "${API_URL}${path}" \
    -H "Authorization: Bearer ${TOKEN}"
}

# Read file and escape for JSON
json_escape() {
  python3 -c "import json,sys; print(json.dumps(sys.stdin.read()))" < "$1"
}

# ── Sync a single consumer-provider pair ──────────────────
sync_pair() {
  local consumer_name="$1"
  local provider_repo="$2"
  local provider_github_id="$3"
  local schema_type="$4"
  local spec_path="$5"
  local spec_file="$6"

  local consumer_github_id
  consumer_github_id=$(echo -n "${ORG}/${consumer_name}" | cksum | awk '{print $1}')

  local payload
  payload=$(jq -n \
    --argjson iid "$INSTALLATION_ID" \
    --arg org "$ORG" \
    --arg consumer "${ORG}/${consumer_name}" \
    --argjson cid "$consumer_github_id" \
    --arg sha "abc123def456" \
    --arg prov "$provider_repo" \
    --argjson pid "$provider_github_id" \
    --arg st "$schema_type" \
    --arg sp "$spec_path" \
    --arg raw "$(cat "$spec_file")" \
    '{
      installation_id: $iid,
      org: $org,
      consumer_repo: $consumer,
      consumer_github_repo_id: $cid,
      commit_sha: $sha,
      dependencies: [{
        provider_repo: $prov,
        provider_github_repo_id: $pid,
        schema_type: $st,
        spec_path: $sp,
        branch: "main",
        raw_content: $raw
      }]
    }')

  if [ "$VERBOSE" = true ]; then
    echo "    Syncing ${consumer_name} -> ${provider_repo}"
  fi

  api_post "/api/v1/sync" "$payload"
}

# ── Sync all demo repos ───────────────────────────────────
test_sync_protobuf() {
  log_section "Sync: microservices-demo (protobuf, 4 consumers)"

  local provider_id=9000001

  for consumer in frontend checkoutservice recommendationservice emailservice; do
    local consumer_id
    consumer_id=$(echo -n "${ORG}/${consumer}" | cksum | awk '{print $1}')

    local payload
    payload=$(jq -n \
      --argjson iid "$INSTALLATION_ID" \
      --arg org "$ORG" \
      --arg consumer "${ORG}/${consumer}" \
      --argjson cid "$consumer_id" \
      --arg sha "abc123def456" \
      --arg prov "${ORG}/microservices-demo" \
      --argjson pid "$provider_id" \
      --arg raw "$(cat "${PAYLOAD_DIR}/protobuf/base.proto")" \
      '{
        installation_id: $iid,
        org: $org,
        consumer_repo: $consumer,
        consumer_github_repo_id: $cid,
        commit_sha: $sha,
        dependencies: [{
          provider_repo: $prov,
          provider_github_repo_id: $pid,
          schema_type: "protobuf",
          spec_path: "protos/demo.proto",
          branch: "main",
          raw_content: $raw
        }]
      }')

    local resp
    resp=$(api_post "/api/v1/sync" "$payload")
    local status
    status=$(echo "$resp" | jq -r '.status // .synced // empty')

    if [ "$status" = "queued" ] || [ "$status" -gt 0 ] 2>/dev/null; then
      log_pass "Synced consumer: ${consumer}"
    else
      log_fail "Sync consumer: ${consumer}" "response: ${resp}"
    fi
  done
}

test_sync_graphql() {
  log_section "Sync: graphql-schema (graphql)"

  local resp
  resp=$(sync_pair "graphql-consumer" "${ORG}/graphql-schema" 9000002 "graphql" "schema.graphql" "${PAYLOAD_DIR}/graphql/base.graphql")
  local status
  status=$(echo "$resp" | jq -r '.status // .synced // empty')

  if [ "$status" = "queued" ] || [ "$status" -gt 0 ] 2>/dev/null; then
    log_pass "Synced graphql-schema consumer"
  else
    log_fail "Sync graphql-schema" "response: ${resp}"
  fi
}

test_sync_openapi() {
  log_section "Sync: stripe/openapi (openapi)"

  local resp
  resp=$(sync_pair "stripe-consumer" "${ORG}/stripe-openapi" 9000003 "openapi" "openapi.yaml" "${PAYLOAD_DIR}/openapi/base.yaml")
  local status
  status=$(echo "$resp" | jq -r '.status // .synced // empty')

  if [ "$status" = "queued" ] || [ "$status" -gt 0 ] 2>/dev/null; then
    log_pass "Synced stripe-openapi consumer"
  else
    log_fail "Sync stripe-openapi" "response: ${resp}"
  fi
}

test_sync_openai() {
  log_section "Sync: openai-openapi (openapi)"

  local resp
  resp=$(sync_pair "openai-consumer" "${ORG}/openai-openapi" 9000004 "openapi" "openapi.yaml" "${PAYLOAD_DIR}/openapi/base.yaml")
  local status
  status=$(echo "$resp" | jq -r '.status // .synced // empty')

  if [ "$status" = "queued" ] || [ "$status" -gt 0 ] 2>/dev/null; then
    log_pass "Synced openai-openapi consumer"
  else
    log_fail "Sync openai-openapi" "response: ${resp}"
  fi
}

test_sync_sql() {
  log_section "Sync: jaffle_shop (sql)"

  local resp
  resp=$(sync_pair "jaffle-consumer" "${ORG}/jaffle_shop" 9000005 "sql" "models/customers.sql" "${PAYLOAD_DIR}/sql/base.sql")
  local status
  status=$(echo "$resp" | jq -r '.status // .synced // empty')

  if [ "$status" = "queued" ] || [ "$status" -gt 0 ] 2>/dev/null; then
    log_pass "Synced jaffle_shop consumer"
  else
    log_fail "Sync jaffle_shop" "response: ${resp}"
  fi
}

test_sync_asyncapi() {
  log_section "Sync: slack-api-specs (asyncapi)"

  local resp
  resp=$(sync_pair "slack-consumer" "${ORG}/slack-api-specs" 9000006 "asyncapi" "events-api/slack_events_api_async_v1.json" "${PAYLOAD_DIR}/asyncapi/base.json")
  local status
  status=$(echo "$resp" | jq -r '.status // .synced // empty')

  if [ "$status" = "queued" ] || [ "$status" -gt 0 ] 2>/dev/null; then
    log_pass "Synced slack-api-specs consumer"
  else
    log_fail "Sync slack-api-specs" "response: ${resp}"
  fi
}

# ── Verify the graph ──────────────────────────────────────
test_graph_verification() {
  log_section "Graph Verification"

  # Wait for River async jobs to process (sync is queued, not synchronous)
  echo "  Waiting for async sync jobs to complete..."
  local attempts=0
  local max_attempts=10
  while [ "$attempts" -lt "$max_attempts" ]; do
    local dep_count
    dep_count=$(api_get "/api/v1/graph/${ORG}" | jq 'length' 2>/dev/null || echo "0")
    if [ "$dep_count" -gt 0 ]; then
      break
    fi
    sleep 1
    attempts=$((attempts + 1))
  done

  local graph
  graph=$(api_get "/api/v1/graph/${ORG}")

  if [ -z "$graph" ] || [ "$graph" = "null" ] || [ "$graph" = "[]" ]; then
    log_fail "Graph fetch" "empty response"
    return
  fi

  # Graph API returns an array of {consumer, provider, status} objects
  local dep_count
  dep_count=$(echo "$graph" | jq 'length' 2>/dev/null || echo "0")

  if [ "$dep_count" -gt 0 ]; then
    log_pass "Graph has ${dep_count} dependency edges"
  else
    log_fail "Graph dependencies" "expected >0, got ${dep_count}"
  fi

  # Check that consumer names are non-empty (validates JSON tags fix)
  local empty_names
  empty_names=$(echo "$graph" | jq -r '.[] | select(.consumer == "" or .consumer == null) | .consumer' 2>/dev/null | wc -l)

  if [ "$empty_names" -eq 0 ]; then
    log_pass "All dependency consumer names are non-empty"
  else
    log_fail "Consumer names" "${empty_names} dependencies have empty consumer names"
  fi

  # Count unique repos (consumers + providers)
  local unique_repos
  unique_repos=$(echo "$graph" | jq -r '.[].consumer, .[].provider' 2>/dev/null | sort -u | wc -l)

  if [ "$unique_repos" -gt 0 ]; then
    log_pass "Graph references ${unique_repos} unique repositories"
  else
    log_fail "Unique repos" "expected >0, got ${unique_repos}"
  fi

  if [ "$VERBOSE" = true ]; then
    echo "    Dependencies:"
    echo "$graph" | jq -r '.[] | "      - \(.consumer) -> \(.provider) [\(.status)]"' 2>/dev/null || true
  fi
}

# ── Run Tests ─────────────────────────────────────────────
echo -e "${YELLOW}Layer 2: API Sync + Graph Verification${NC}"
echo "API URL: ${API_URL}"
echo "Org: ${ORG}"

test_sync_protobuf
test_sync_graphql
test_sync_openapi
test_sync_openai
test_sync_sql
test_sync_asyncapi
test_graph_verification

echo -e "\n${YELLOW}━━━ Results ━━━${NC}"
echo -e "  Total: ${TOTAL}  ${GREEN}Passed: ${PASS}${NC}  ${RED}Failed: ${FAIL}${NC}"

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
