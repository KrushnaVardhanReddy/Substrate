#!/usr/bin/env bash
# test-engine.sh — Layer 1: Engine Diff Tests
# Tests the Go diff engine directly for all supported schema types.
# Usage: ./scripts/test-engine.sh [--schema=protobuf|graphql|openapi|sql|asyncapi|all] [--verbose]

set -euo pipefail

ENGINE_URL="${ENGINE_URL:-http://localhost:8080}"
PAYLOAD_DIR="$(dirname "$0")/test-payloads"
SCHEMA_FILTER="all"
VERBOSE=false

for arg in "$@"; do
  case "$arg" in
    --schema=*) SCHEMA_FILTER="${arg#*=}" ;;
    --verbose) VERBOSE=true ;;
  esac
done

PASS=0
FAIL=0
TOTAL=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_pass() { PASS=$((PASS + 1)); TOTAL=$((TOTAL + 1)); echo -e "  ${GREEN}✓ PASS${NC} $1"; }
log_fail() { FAIL=$((FAIL + 1)); TOTAL=$((TOTAL + 1)); echo -e "  ${RED}✗ FAIL${NC} $1: $2"; }
log_section() { echo -e "\n${YELLOW}━━━ $1 ━━━${NC}"; }

# Call the engine /diff endpoint using jq for safe JSON construction
call_engine() {
  local schema_type="$1"
  local base_file="$2"
  local head_file="$3"

  local payload
  payload=$(jq -n \
    --arg base "$(cat "$base_file")" \
    --arg head "$(cat "$head_file")" \
    --arg st "$schema_type" \
    '{base_schema: $base, head_schema: $head, schema_type: $st}')

  if [ "$VERBOSE" = true ]; then
    echo "    Request payload (truncated): ${payload:0:200}..."
  fi

  curl -s -X POST "${ENGINE_URL}/diff" \
    -H 'Content-Type: application/json' \
    -d "$payload"
}

# ── Protobuf ──────────────────────────────────────────────
test_protobuf() {
  log_section "Protobuf (microservices-demo)"

  local base="${PAYLOAD_DIR}/protobuf/base.proto"
  local head_br="${PAYLOAD_DIR}/protobuf/head-breaking.proto"
  local head_safe="${PAYLOAD_DIR}/protobuf/head-safe.proto"

  # Breaking: removed product_id field
  local resp
  resp=$(call_engine "protobuf" "$base" "$head_br")
  local breaking_count
  breaking_count=$(echo "$resp" | jq -r '.summary.breaking_count // 0')
  local has_rule
  has_rule=$(echo "$resp" | jq -r '.breaking_changes[]? | select(.rule_id | test("PROTO")) | .rule_id' | head -1)

  if [ "$breaking_count" -gt 0 ] && [ -n "$has_rule" ]; then
    log_pass "Breaking change detected (${has_rule})"
  else
    log_fail "Breaking change" "expected breaking_count>0, got ${breaking_count}"
  fi

  # Safe: added notes field
  resp=$(call_engine "protobuf" "$base" "$head_safe")
  breaking_count=$(echo "$resp" | jq -r '.summary.breaking_count // 0')

  if [ "$breaking_count" -eq 0 ]; then
    log_pass "Safe extension accepted"
  else
    log_fail "Safe extension" "expected breaking_count=0, got ${breaking_count}"
  fi
}

# ── GraphQL ───────────────────────────────────────────────
test_graphql() {
  log_section "GraphQL (github/graphql-schema)"

  local base="${PAYLOAD_DIR}/graphql/base.graphql"
  local head_br="${PAYLOAD_DIR}/graphql/head-breaking.graphql"
  local head_safe="${PAYLOAD_DIR}/graphql/head-safe.graphql"

  # Breaking: removed email field
  local resp
  resp=$(call_engine "graphql" "$base" "$head_br")
  local breaking_count
  breaking_count=$(echo "$resp" | jq -r '.summary.breaking_count // 0')
  local has_rule
  has_rule=$(echo "$resp" | jq -r '.breaking_changes[]? | select(.rule_id | test("GQL")) | .rule_id' | head -1)

  if [ "$breaking_count" -gt 0 ] && [ -n "$has_rule" ]; then
    log_pass "Breaking change detected (${has_rule})"
  else
    log_fail "Breaking change" "expected breaking_count>0, got ${breaking_count}"
  fi

  # Safe: added age field
  resp=$(call_engine "graphql" "$base" "$head_safe")
  breaking_count=$(echo "$resp" | jq -r '.summary.breaking_count // 0')

  if [ "$breaking_count" -eq 0 ]; then
    log_pass "Safe extension accepted"
  else
    log_fail "Safe extension" "expected breaking_count=0, got ${breaking_count}"
  fi
}

# ── OpenAPI ───────────────────────────────────────────────
test_openapi() {
  log_section "OpenAPI (stripe/openapi + openai-openapi)"

  local base="${PAYLOAD_DIR}/openapi/base.yaml"
  local head_br="${PAYLOAD_DIR}/openapi/head-breaking.yaml"
  local head_safe="${PAYLOAD_DIR}/openapi/head-safe.yaml"

  # Breaking: removed /users/{id} endpoint
  local resp
  resp=$(call_engine "openapi" "$base" "$head_br")
  local breaking_count
  breaking_count=$(echo "$resp" | jq -r '.summary.breaking_count // 0')
  local has_rule
  has_rule=$(echo "$resp" | jq -r '.breaking_changes[]? | select(.rule_id | test("ENDPOINT")) | .rule_id' | head -1)

  if [ "$breaking_count" -gt 0 ] && [ -n "$has_rule" ]; then
    log_pass "Breaking change detected (${has_rule})"
  else
    log_fail "Breaking change" "expected breaking_count>0, got ${breaking_count}"
  fi

  # Safe: added /users/{id}/profile endpoint
  resp=$(call_engine "openapi" "$base" "$head_safe")
  breaking_count=$(echo "$resp" | jq -r '.summary.breaking_count // 0')

  if [ "$breaking_count" -eq 0 ]; then
    log_pass "Safe extension accepted"
  else
    log_fail "Safe extension" "expected breaking_count=0, got ${breaking_count}"
  fi
}

# ── SQL ───────────────────────────────────────────────────
test_sql() {
  log_section "SQL (dbt-labs/jaffle_shop)"

  local base="${PAYLOAD_DIR}/sql/base.sql"
  local head_br="${PAYLOAD_DIR}/sql/head-breaking.sql"
  local head_safe="${PAYLOAD_DIR}/sql/head-safe.sql"

  # Breaking: dropped email column
  local resp
  resp=$(call_engine "sql" "$base" "$head_br")
  local breaking_count
  breaking_count=$(echo "$resp" | jq -r '.summary.breaking_count // 0')
  local has_rule
  has_rule=$(echo "$resp" | jq -r '.breaking_changes[]? | select(.rule_id | test("COLUMN")) | .rule_id' | head -1)

  if [ "$breaking_count" -gt 0 ] && [ -n "$has_rule" ]; then
    log_pass "Breaking change detected (${has_rule})"
  else
    log_fail "Breaking change" "expected breaking_count>0, got ${breaking_count}"
  fi

  # Safe: added nullable age column
  resp=$(call_engine "sql" "$base" "$head_safe")
  breaking_count=$(echo "$resp" | jq -r '.summary.breaking_count // 0')

  if [ "$breaking_count" -eq 0 ]; then
    log_pass "Safe extension accepted"
  else
    log_fail "Safe extension" "expected breaking_count=0, got ${breaking_count}"
  fi
}

# ── AsyncAPI ──────────────────────────────────────────────
test_asyncapi() {
  log_section "AsyncAPI (slackapi/slack-api-specs)"

  local base="${PAYLOAD_DIR}/asyncapi/base.json"
  local head_br="${PAYLOAD_DIR}/asyncapi/head-breaking.json"
  local head_safe="${PAYLOAD_DIR}/asyncapi/head-safe.json"

  # Breaking: removed channel from required
  local resp
  resp=$(call_engine "asyncapi" "$base" "$head_br")
  local breaking_count
  breaking_count=$(echo "$resp" | jq -r '.summary.breaking_count // 0')

  if [ "$breaking_count" -gt 0 ]; then
    log_pass "Breaking change detected"
  else
    log_fail "Breaking change" "expected breaking_count>0, got ${breaking_count}"
  fi

  # Safe: added optional thread_ts
  resp=$(call_engine "asyncapi" "$base" "$head_safe")
  breaking_count=$(echo "$resp" | jq -r '.summary.breaking_count // 0')

  if [ "$breaking_count" -eq 0 ]; then
    log_pass "Safe extension accepted"
  else
    log_fail "Safe extension" "expected breaking_count=0, got ${breaking_count}"
  fi
}

# ── Run Tests ─────────────────────────────────────────────
echo -e "${YELLOW}Layer 1: Engine Diff Tests${NC}"
echo "Engine URL: ${ENGINE_URL}"

case "$SCHEMA_FILTER" in
  protobuf) test_protobuf ;;
  graphql)  test_graphql ;;
  openapi)  test_openapi ;;
  sql)      test_sql ;;
  asyncapi) test_asyncapi ;;
  all)
    test_protobuf
    test_graphql
    test_openapi
    test_sql
    test_asyncapi
    ;;
  *)
    echo "Unknown schema: ${SCHEMA_FILTER}"
    exit 2
    ;;
esac

echo -e "\n${YELLOW}━━━ Results ━━━${NC}"
echo -e "  Total: ${TOTAL}  ${GREEN}Passed: ${PASS}${NC}  ${RED}Failed: ${FAIL}${NC}"

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
