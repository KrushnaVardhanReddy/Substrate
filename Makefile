.PHONY: help e2e e2e-breaking e2e-safe e2e-override e2e-warning postgres api engine worker dashboard docs build-cli build-mcp start-bg stop-bg

# ==============================================================================
# SUBSTRATE LOCAL DEVELOPMENT ARCHITECTURE
#
# Substrate is built as a highly decoupled, microservice-like architecture to 
# allow maximum flexibility (CLI-mode vs Cloud-mode vs IDE-mode).
#
# To run Substrate locally, you need 5 terminals running simultaneously:
#
# 1. Database (make postgres)
#    - Runs a local PostgreSQL 15 container to store the cross-repo graph.
# 2. Registry API (cd api && go run ./cmd/server)
#    - The stateful brain. Connects to Postgres, maps dependencies, and 
#      exposes the graph to the Dashboard and MCP.
# 3. Diff Engine (cd engine && go run ./cmd/substrate serve)
#    - The stateless worker. It only takes two schemas, compares them, and 
#      returns the breaking changes. Exposed over port 8080.
# 4. GitHub Worker (cd github-app && npm run dev)
#    - The Cloudflare Worker that listens to GitHub Webhooks, fetches the 
#      PR files, and orchestrates the Registry API and Diff Engine.
# 5. Svelte Dashboard (cd dashboard && npm run dev)
#    - The visualizer UI to see the live graph and audit history.
#
# Why multiple Golang binaries?
# - engine/cmd/substrate: A stateless CLI tool that can be run in Github Actions 
#   or as a microservice (serve).
# - api/cmd/server: A stateful API that requires Postgres. Separated from the 
#   engine so the engine can be used purely locally/offline.
# - engine/cmd/substrate-mcp: A specialized wrapper that runs the engine 
#   functions over standard input/output (stdio) using the JSON-RPC Model 
#   Context Protocol for AI IDEs like Cursor and Claude.
# ==============================================================================

help:
	@echo "Substrate Local Development Commands:"
	@echo "--------------------------------------------------------"
	@echo "make postgres     - Start the Postgres database in Docker"
	@echo "make api          - Start the Registry API (port 8090)"
	@echo "make engine       - Start the Diff Engine (port 8080)"
	@echo "make worker       - Start the GitHub Webhook Worker"
	@echo "make dashboard    - Start the Svelte Dashboard UI"
	@echo "make docs         - Start the Astro Starlight Docs site"
	@echo "make build-cli    - Build the substrate CLI binary"
	@echo "make build-mcp    - Build the substrate-mcp binary"
	@echo "--------------------------------------------------------"
	@echo "make start-bg     - Start ALL backend services in background"
	@echo "make stop-bg      - Stop all background backend services"
	@echo "make e2e-*        - Run the E2E matrix test suites"
	@echo "--------------------------------------------------------"
	@echo "See the Makefile source for the full 5-terminal architecture setup."

postgres:
	podman run --replace --name substrate-postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=substrate -p 5432:5432 -d docker.io/library/postgres:15

api:
	cd api && \
	DATABASE_URL="postgresql://postgres:postgres@localhost:5432/substrate?sslmode=disable" \
	REGISTRY_API_TOKEN="local-dev-token" \
	JWT_SECRET="local-jwt-secret" \
	GITHUB_CLIENT_ID="mock-client-id" \
	GITHUB_CLIENT_SECRET="mock-client-secret" \
	DASHBOARD_URL="http://localhost:5173" \
	go run ./cmd/server/main.go

# To enable real AI (requires LM Studio running at port 1234), use make api-ai instead
api-ai:
	cd api && \
	DATABASE_URL="postgresql://postgres:postgres@localhost:5432/substrate?sslmode=disable" \
	REGISTRY_API_TOKEN="local-dev-token" \
	JWT_SECRET="local-jwt-secret" \
	GITHUB_CLIENT_ID="mock-client-id" \
	GITHUB_CLIENT_SECRET="mock-client-secret" \
	DASHBOARD_URL="http://localhost:5173" \
	SUBSTRATE_AI_BASE_URL="http://127.0.0.1:1234/v1" \
	SUBSTRATE_AI_API_KEY="lm-studio" \
	SUBSTRATE_AI_MODEL="qwen/qwen3.5-9b" \
	go run ./cmd/server/main.go

engine:
	cd engine && go run ./cmd/substrate/ serve

worker:
	cd github-app && npm run dev

dashboard:
	cd dashboard && npm run dev

docs:
	cd docs-site && npm run dev

build-cli:
	cd engine && go build -o substrate ./cmd/substrate/

build-mcp:
	cd engine && go build -o substrate-mcp ./cmd/substrate-mcp/main.go

start-bg: postgres
	@echo "Starting backend services in background..."
	@make api > api.log 2>&1 & echo $$! > api.pid
	@make engine > engine.log 2>&1 & echo $$! > engine.pid
	@make worker > worker.log 2>&1 & echo $$! > worker.pid
	@make dashboard > dashboard.log 2>&1 & echo $$! > dashboard.pid
	@echo "Starting ngrok tunnel for webhook routing..."
	@ngrok http 8787 > ngrok.log 2>&1 & echo $$! > ngrok.pid
	@sleep 5
	@echo "=========================================================="
	@echo "✅ Services started. Logs available in api.log, engine.log, worker.log, dashboard.log"
	@echo "⚠️ ACTION REQUIRED: Update your GitHub App Webhook URL to:"
	@curl -s http://localhost:4040/api/tunnels | grep -o '"public_url":"https://[^"]*"' | cut -d'"' -f4 || echo "Failed to fetch ngrok URL (check ngrok.log)"
	@echo "=========================================================="
	@echo "Run 'make stop-bg' to terminate."

stop-bg:
	@echo "Stopping backend services..."
	@-kill `cat api.pid` 2>/dev/null || true
	@-kill `cat engine.pid` 2>/dev/null || true
	@-kill `cat worker.pid` 2>/dev/null || true
	@-kill `cat dashboard.pid` 2>/dev/null || true
	@-kill `cat ngrok.pid` 2>/dev/null || true
	@-fuser -k 8080/tcp 2>/dev/null || true
	@-fuser -k 8090/tcp 2>/dev/null || true
	@-fuser -k 5173/tcp 2>/dev/null || true
	@rm -f api.pid engine.pid worker.pid dashboard.pid ngrok.pid api.log engine.log worker.log dashboard.log ngrok.log
	@podman stop substrate-postgres || true

# Ensure GITHUB_TOKEN is set before running these
check-token:
	@if [ -z "$(GITHUB_TOKEN)" ]; then \
		echo "Error: GITHUB_TOKEN is not set."; \
		echo "Export it using: export GITHUB_TOKEN=ghp_..."; \
		exit 1; \
	fi

e2e: check-token
	cd scripts/e2e && go run main.go --scenario=all

e2e-openapi: check-token
	cd scripts/e2e && go run main.go --scenario=openapi

e2e-sql: check-token
	cd scripts/e2e && go run main.go --scenario=sql

e2e-graphql: check-token
	cd scripts/e2e && go run main.go --scenario=graphql

e2e-protobuf: check-token
	cd scripts/e2e && go run main.go --scenario=protobuf


e2e-breaking: check-token
	cd scripts/e2e && go run main.go --scenario=openapi-breaking

e2e-safe: check-token
	cd scripts/e2e && go run main.go --scenario=openapi-safe

e2e-override: check-token
	cd scripts/e2e && go run main.go --scenario=openapi-override

e2e-warning: check-token
	cd scripts/e2e && go run main.go --scenario=openapi-warning

e2e-sql-breaking: check-token
	cd scripts/e2e && go run main.go --scenario=sql-breaking

e2e-sql-safe: check-token
	cd scripts/e2e && go run main.go --scenario=sql-safe

e2e-sql-override: check-token
	cd scripts/e2e && go run main.go --scenario=sql-override

e2e-sql-warning: check-token
	cd scripts/e2e && go run main.go --scenario=sql-warning

e2e-graphql-breaking: check-token
	cd scripts/e2e && go run main.go --scenario=graphql-breaking

e2e-graphql-safe: check-token
	cd scripts/e2e && go run main.go --scenario=graphql-safe

e2e-graphql-override: check-token
	cd scripts/e2e && go run main.go --scenario=graphql-override

e2e-graphql-warning: check-token
	cd scripts/e2e && go run main.go --scenario=graphql-warning


e2e-protobuf-breaking: check-token
	cd scripts/e2e && go run main.go --scenario=protobuf-breaking

e2e-protobuf-safe: check-token
	cd scripts/e2e && go run main.go --scenario=protobuf-safe

e2e-protobuf-override: check-token
	cd scripts/e2e && go run main.go --scenario=protobuf-override

e2e-protobuf-warning: check-token
	cd scripts/e2e && go run main.go --scenario=protobuf-warning

e2e-asyncapi: check-token
	cd scripts/e2e && go run main.go --scenario=asyncapi

e2e-asyncapi-breaking: check-token
	cd scripts/e2e && go run main.go --scenario=asyncapi-breaking

e2e-asyncapi-safe: check-token
	cd scripts/e2e && go run main.go --scenario=asyncapi-safe

e2e-asyncapi-override: check-token
	cd scripts/e2e && go run main.go --scenario=asyncapi-override

e2e-asyncapi-warning: check-token
	cd scripts/e2e && go run main.go --scenario=asyncapi-warning

e2e-avro: check-token
	cd scripts/e2e && go run main.go --scenario=avro

e2e-avro-breaking: check-token
	cd scripts/e2e && go run main.go --scenario=avro-breaking

e2e-avro-safe: check-token
	cd scripts/e2e && go run main.go --scenario=avro-safe

e2e-terraform: check-token
	cd scripts/e2e && go run main.go --scenario=terraform

e2e-terraform-breaking: check-token
	cd scripts/e2e && go run main.go --scenario=terraform-breaking

e2e-terraform-safe: check-token
	cd scripts/e2e && go run main.go --scenario=terraform-safe

e2e-terraform-override: check-token
	cd scripts/e2e && go run main.go --scenario=terraform-override

e2e-aiml: check-token
	cd scripts/e2e && go run main.go --scenario=aiml

e2e-aiml-breaking: check-token
	cd scripts/e2e && go run main.go --scenario=aiml-breaking

e2e-aiml-safe: check-token
	cd scripts/e2e && go run main.go --scenario=aiml-safe

e2e-aiml-override: check-token
	cd scripts/e2e && go run main.go --scenario=aiml-override

e2e-discovery:
	@echo "Running Phase 5 Cross-Repo Dependency Discovery E2E Tests..."
	cd api && go test -v -run TestPhase5DependencyDiscoveryE2E ./internal/discovery

e2e-scale: check-token
	cd scripts/e2e && go run scale_generator.go --scale=100

e2e-v1: check-token
	@echo "Running V1.0 System E2E Tests..."
	cd scripts/e2e && go test -v v1_e2e_test.go
