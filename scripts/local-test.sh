#!/usr/bin/env bash
# local-test.sh — Main orchestrator for local demo repository testing
# Runs all 3 layers: Engine Diff → API Sync → Cross-Repo Blast Radius
# No GitHub token required. Bypasses Cloudflare Worker entirely.
#
# Usage:
#   ./scripts/local-test.sh                         # Run all layers, all schemas
#   ./scripts/local-test.sh --layer=1               # Engine diff only
#   ./scripts/local-test.sh --schema=protobuf       # Protobuf only
#   ./scripts/local-test.sh --reset                 # Reset DB, then test
#   ./scripts/local-test.sh --verbose               # Print full response bodies
#
# Prerequisites:
#   - make start-bg running (Postgres, API, Engine, Worker, Dashboard)
#   - curl and jq installed

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ENGINE_URL="${ENGINE_URL:-http://localhost:8080}"
API_URL="${API_URL:-http://localhost:8090}"
LAYER="all"
SCHEMA="all"
VERBOSE=false
RESET=false

for arg in "$@"; do
  case "$arg" in
    --layer=*) LAYER="${arg#*=}" ;;
    --schema=*) SCHEMA="${arg#*=}" ;;
    --verbose) VERBOSE=true ;;
    --reset) RESET=true ;;
    --help|-h)
      echo "Usage: $0 [--layer=1|2|3|all] [--schema=protobuf|graphql|openapi|sql|asyncapi|all] [--reset] [--verbose]"
      exit 0
      ;;
  esac
done

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}╔══════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║  Substrate Local Demo Testing Suite                     ║${NC}"
echo -e "${CYAN}║  No GitHub token required. Bypasses Cloudflare Worker.  ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════════════════════╝${NC}"
echo ""

# ── Prerequisite Checks ──────────────────────────────────
echo -e "${YELLOW}Checking prerequisites...${NC}"

check_service() {
  local name="$1"
  local url="$2"
  local expected="$3"

  local resp
  resp=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null || echo "000")

  if [ "$resp" = "$expected" ] || ([ "$expected" = "2xx" ] && [[ "$resp" =~ ^2 ]]); then
    echo -e "  ${GREEN}✓${NC} ${name} (HTTP ${resp})"
    return 0
  else
    echo -e "  ${RED}✗${NC} ${name} (HTTP ${resp} — expected ${expected})"
    return 1
  fi
}

PREREQS_OK=true

check_service "Go API" "${API_URL}/health" "200" || PREREQS_OK=false
# Engine has no root health endpoint; check by hitting /diff with empty body (expects 400)
ENGINE_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${ENGINE_URL}/diff" -H 'Content-Type: application/json' -d '{}' 2>/dev/null || echo "000")
if [ "$ENGINE_STATUS" = "400" ] || [ "$ENGINE_STATUS" = "200" ]; then
  echo -e "  ${GREEN}✓${NC} Go Engine (HTTP ${ENGINE_STATUS} — running)"
else
  echo -e "  ${RED}✗${NC} Go Engine (HTTP ${ENGINE_STATUS} — not responding)"
  PREREQS_OK=false
fi

# Check Postgres container
if command -v docker &>/dev/null; then
  PG_STATUS=$(docker ps --filter name=substrate-postgres --format "{{.Status}}" 2>/dev/null || echo "")
  if [ -n "$PG_STATUS" ]; then
    echo -e "  ${GREEN}✓${NC} PostgreSQL (${PG_STATUS})"
  else
    echo -e "  ${RED}✗${NC} PostgreSQL (container not running)"
    PREREQS_OK=false
  fi
elif command -v podman &>/dev/null; then
  PG_STATUS=$(podman ps --filter name=substrate-postgres --format "{{.Status}}" 2>/dev/null || echo "")
  if [ -n "$PG_STATUS" ]; then
    echo -e "  ${GREEN}✓${NC} PostgreSQL (${PG_STATUS})"
  else
    echo -e "  ${RED}✗${NC} PostgreSQL (container not running)"
    PREREQS_OK=false
  fi
fi

# Check jq
if command -v jq &>/dev/null; then
  echo -e "  ${GREEN}✓${NC} jq $(jq --version 2>/dev/null)"
else
  echo -e "  ${RED}✗${NC} jq (not installed — required for JSON parsing)"
  PREREQS_OK=false
fi

if [ "$PREREQS_OK" = false ]; then
  echo -e "\n${RED}Prerequisites failed. Run 'make start-bg' first.${NC}"
  exit 2
fi

echo -e "\n${GREEN}All prerequisites met.${NC}"

# ── Optional: Reset Database ──────────────────────────────
if [ "$RESET" = true ]; then
  echo -e "\n${YELLOW}Resetting database...${NC}"
  make reset-demo 2>/dev/null || {
    echo "  Falling back to manual reset..."
    docker rm -f substrate-postgres 2>/dev/null || podman rm -f substrate-postgres 2>/dev/null || true
    make postgres 2>/dev/null
    sleep 3
  }
  echo -e "${GREEN}Database reset complete.${NC}"
  echo -e "${YELLOW}Note: You may need to restart services with 'make start-bg'${NC}"
fi

# ── Export variables for sub-scripts ──────────────────────
export ENGINE_URL
export API_URL
export VERBOSE

# ── Run Layers ────────────────────────────────────────────
OVERALL_EXIT=0

run_layer() {
  local layer_num="$1"
  local layer_name="$2"
  local script="$3"

  echo -e "\n${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
  echo -e "${CYAN}  Layer ${layer_num}: ${layer_name}${NC}"
  echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

  local args=()
  if [ "$SCHEMA" != "all" ]; then
    args+=("--schema=${SCHEMA}")
  fi
  if [ "$VERBOSE" = true ]; then
    args+=("--verbose")
  fi

  if bash "${SCRIPT_DIR}/${script}" "${args[@]}"; then
    echo -e "${GREEN}Layer ${layer_num} complete.${NC}"
  else
    echo -e "${RED}Layer ${layer_num} had failures.${NC}"
    OVERALL_EXIT=1
  fi
}

START_TIME=$(date +%s)

case "$LAYER" in
  1)    run_layer 1 "Engine Diff Tests" "test-engine.sh" ;;
  2)    run_layer 2 "API Sync + Graph" "test-sync.sh" ;;
  3)    run_layer 3 "Cross-Repo Blast Radius" "test-cross-repo.sh" ;;
  all)
    run_layer 1 "Engine Diff Tests" "test-engine.sh"
    run_layer 2 "API Sync + Graph" "test-sync.sh"
    run_layer 3 "Cross-Repo Blast Radius" "test-cross-repo.sh"
    ;;
  *)
    echo "Unknown layer: ${LAYER}. Use 1, 2, 3, or all."
    exit 2
    ;;
esac

END_TIME=$(date +%s)
ELAPSED=$((END_TIME - START_TIME))

# ── Final Summary ─────────────────────────────────────────
echo ""
echo -e "${CYAN}╔══════════════════════════════════════════════════════════╗${NC}"
if [ "$OVERALL_EXIT" -eq 0 ]; then
  echo -e "${CYAN}║  ${GREEN}ALL TESTS PASSED${CYAN}                                      ║${NC}"
else
  echo -e "${CYAN}║  ${RED}SOME TESTS FAILED${CYAN}                                     ║${NC}"
fi
echo -e "${CYAN}║  Completed in ${ELAPSED}s                                    ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════════════════════╝${NC}"

exit "$OVERALL_EXIT"
