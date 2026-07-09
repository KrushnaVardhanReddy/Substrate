.PHONY: e2e e2e-breaking e2e-safe e2e-override e2e-warning

# Ensure GITHUB_TOKEN is set before running these
check-token:
	@if [ -z "$(GITHUB_TOKEN)" ]; then \
		echo "Error: GITHUB_TOKEN is not set."; \
		echo "Export it using: export GITHUB_TOKEN=ghp_..."; \
		exit 1; \
	fi

e2e: check-token
	cd scripts/e2e && go run main.go --scenario=all

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
