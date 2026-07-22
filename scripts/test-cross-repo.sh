#!/usr/bin/env bash
# test-cross-repo.sh — Layer 3: Cross-Repo Blast Radius Verification
# After syncing, tests cross-repo impact analysis for breaking changes.
# Usage: ./scripts/test-cross-repo.sh [--verbose]

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

json_escape() {
  python3 -c "import json,sys; print(json.dumps(sys.stdin.read()))" < "$1"
}

# ── Cross-repo check for protobuf (microservices-demo) ────
test_cross_repo_protobuf() {
  log_section "Cross-Repo: microservices-demo (4 downstream consumers)"

  local payload
  payload=$(jq -n \
    --argjson iid "$INSTALLATION_ID" \
    --arg org "$ORG" \
    --arg prov "${ORG}/microservices-demo" \
    --arg raw "$(cat "${PAYLOAD_DIR}/protobuf/head-breaking.proto")" \
    --arg st "protobuf" \
    --arg cfg "schema_type: protobuf\nbase_schema: protos/demo.proto\nhead_schema: protos/demo.proto" \
    '{
      installation_id: $iid,
      org: $org,
      provider_repo: $prov,
      head_schema_content: $raw,
      schema_type: $st,
      config_content: $cfg
    }')

  local resp
  resp=$(curl -s -X POST "${API_URL}/api/v1/cross-repo-check" \
    -H "Authorization: Bearer ${TOKEN}" \
    -H 'Content-Type: application/json' \
    -d "$payload")

  if [ "$VERBOSE" = true ]; then
    echo "    Response: ${resp}"
  fi

  # Check for 402 (tier limit) — should NOT happen in development mode
  if echo "$resp" | grep -q "402\|Payment Required\|Free tier limit"; then
    log_fail "Tier limit bypass" "got 402 Payment Required — ENVIRONMENT=development not set?"
    return
  fi

  # The cross-repo-check endpoint enqueues an async job and returns 202 {"status":"queued"}
  local status
  status=$(echo "$resp" | jq -r '.status // empty')

  if [ "$status" = "queued" ]; then
    log_pass "Cross-repo check accepted (async job queued)"
    log_pass "Tier limits bypassed (no 402 in development mode)"
    return
  fi

  # If the response has actual results (job completed synchronously), validate them
  local total_consumers
  total_consumers=$(echo "$resp" | jq -r '.total_consumers // 0')
  local broken_consumers
  broken_consumers=$(echo "$resp" | jq -r '.broken_consumers // 0')
  local is_safe
  is_safe=$(echo "$resp" | jq -r '.is_safe // true')

  if [ "$total_consumers" -gt 0 ]; then
    log_pass "Found ${total_consumers} downstream consumer(s)"
  else
    log_fail "Downstream consumers" "expected >0, got ${total_consumers} — did sync run?"
  fi

  if [ "$broken_consumers" -gt 0 ]; then
    log_pass "Detected ${broken_consumers} broken consumer(s) from breaking change"
  else
    log_fail "Broken consumers" "expected >0 with breaking proto change, got ${broken_consumers}"
  fi

  if [ "$is_safe" = "false" ] || [ "$broken_consumers" -gt 0 ]; then
    log_pass "Cross-repo check correctly flagged as unsafe"
  else
    log_fail "Safety flag" "expected is_safe=false with breaking change"
  fi

  # Verify individual consumer results
  local results
  results=$(echo "$resp" | jq -r '.results[]? | "\(.consumer_repo)|\(.is_safe)|\(.rule_id // "none")"' 2>/dev/null)

  if [ -n "$results" ]; then
    local consumer_count
    consumer_count=$(echo "$results" | wc -l)
    log_pass "Detailed results for ${consumer_count} consumer(s)"
    if [ "$VERBOSE" = true ]; then
      echo "$results" | while IFS='|' read -r repo safe rule; do
        echo "      ${repo}: safe=${safe} rule=${rule}"
      done
    fi
  fi
}

# ── Cross-repo check for OpenAPI (single consumer) ────────
test_cross_repo_openapi() {
  log_section "Cross-Repo: stripe-openapi (1 downstream consumer)"

  local payload
  payload=$(jq -n \
    --argjson iid "$INSTALLATION_ID" \
    --arg org "$ORG" \
    --arg prov "${ORG}/stripe-openapi" \
    --arg raw "$(cat "${PAYLOAD_DIR}/openapi/head-breaking.yaml")" \
    --arg st "openapi" \
    --arg cfg "schema_type: openapi\nbase_schema: openapi.yaml\nhead_schema: openapi.yaml" \
    '{
      installation_id: $iid,
      org: $org,
      provider_repo: $prov,
      head_schema_content: $raw,
      schema_type: $st,
      config_content: $cfg
    }')

  local resp
  resp=$(curl -s -X POST "${API_URL}/api/v1/cross-repo-check" \
    -H "Authorization: Bearer ${TOKEN}" \
    -H 'Content-Type: application/json' \
    -d "$payload")

  local status
  status=$(echo "$resp" | jq -r '.status // empty')

  if [ "$status" = "queued" ]; then
    log_pass "Cross-repo check accepted (async job queued)"
    return
  fi

  local total_consumers
  total_consumers=$(echo "$resp" | jq -r '.total_consumers // 0')

  if [ "$total_consumers" -gt 0 ]; then
    log_pass "Found ${total_consumers} downstream consumer(s)"
  else
    log_fail "Downstream consumers" "expected >0, got ${total_consumers}"
  fi
}

# ── Diff report storage ───────────────────────────────────
test_diff_storage() {
  log_section "Diff Report Storage"

  local payload
  payload=$(jq -n \
    --argjson iid "$INSTALLATION_ID" \
    --arg org "$ORG" \
    '{
      diff_report: {
        breaking_changes: [{
          rule_id: "PROTO_FIELD_REMOVED",
          severity: "BREAKING",
          path: "CartItem.product_id",
          description: "Field product_id removed from message CartItem"
        }],
        warnings: [],
        safe_changes: [],
        summary: {
          breaking_count: 1,
          warning_count: 0,
          info_count: 0
        }
      },
      is_audit_mode: false,
      org: $org,
      provider_repo: ($org + "/microservices-demo"),
      pr_number: 1,
      commit_sha: "abc123def456",
      schema_type: "protobuf",
      config_content: "schema_type: protobuf"
    }')

  local resp
  resp=$(curl -s -X POST "${API_URL}/api/v1/diff" \
    -H "Authorization: Bearer ${TOKEN}" \
    -H 'Content-Type: application/json' \
    -d "$payload")

  local diff_id
  diff_id=$(echo "$resp" | jq -r '.id // empty')

  if [ -n "$diff_id" ]; then
    log_pass "Diff report stored with id: ${diff_id}"
  else
    log_fail "Diff storage" "no id returned: ${resp}"
  fi
}

# ── Run Tests ─────────────────────────────────────────────
echo -e "${YELLOW}Layer 3: Cross-Repo Blast Radius${NC}"
echo "API URL: ${API_URL}"
echo "Org: ${ORG}"

test_cross_repo_protobuf
test_cross_repo_openapi
test_diff_storage

echo -e "\n${YELLOW}━━━ Results ━━━${NC}"
echo -e "  Total: ${TOTAL}  ${GREEN}Passed: ${PASS}${NC}  ${RED}Failed: ${FAIL}${NC}"

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
