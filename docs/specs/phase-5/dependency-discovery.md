# Phase 5 Spec: Automated Dependency Discovery

> **Status:** 💡 Planned — Phase 5 (Enterprise)
> **Depends on:** Phase 3 Contract Registry (stores URL→repo mappings)
> **Goal:** Eliminate all manual `substrate.yaml` consumer declarations by automatically building the full cross-repo dependency graph from signals already present in the codebase and infrastructure.

---

## The Core Problem

In Phase 3, a consumer (e.g. `frontend`) must manually declare:

```yaml
# substrate.yaml — written and maintained by hand
consumers:
  - name: "users-api"
    provider_repo: "myorg/backend-api"
```

This works but has friction: developers forget to update it, new services never get added, and it requires onboarding effort for every repo. The goal of Phase 5 is to **make this declaration automatic** — Substrate discovers dependencies by reading signals already present in the codebase.

---

## Discovery Strategy: Multi-Signal, Confidence-Scored

No single discovery method is 100% reliable. Substrate will run **all available scanners in parallel** and combine their results into a confidence score per dependency edge. Teams see not just the dependency, but how certain Substrate is about it.

```
frontend → backend-api

  ✅ Signal: package.json imports @myorg/backend-sdk           +40 pts
  ✅ Signal: .env.example has BACKEND_API_URL=api.myorg.com    +30 pts
  ✅ Signal: docker-compose.yml ENV wires frontend→backend     +20 pts
  ✅ Signal: OTel trace confirms live HTTP call                 +10 pts
                                                            ─────────
                                     Confidence Score:         100 pts → HIGH ✅
```

| Score Range | Confidence Label | UI Treatment |
|---|---|---|
| 80–100 | ✅ **High** | Solid edge in graph, used for breaking change checks |
| 50–79  | ⚠️ **Medium** | Dashed edge, included but flagged for manual review |
| 20–49  | 🔍 **Low** | Hidden by default, shown in "unconfirmed" panel |
| <20    | ❌ **Noise** | Discarded |

---

## Tier 1 — Static In-Repo Signals (No Infrastructure Required)

These scanners run entirely on the repo contents at CI time. No external access needed.

### 1A. Environment Variable Scanning ⭐ (Highest Impact)

**The insight:** Almost every service stores its upstream dependencies as environment variables:
```
USERS_API_URL=https://users.myorg.com
PAYMENTS_ENDPOINT=https://api.payments.myorg.com/v2
```

The variable *name* (`USERS_API_URL`) is a strong signal even without the value. The variable *value* can be resolved against the URL→repo registry to produce a definitive link.

#### Files to scan (safe to commit — no secrets):

| File | Why it's safe | What we extract |
|---|---|---|
| `.env.example` / `.env.template` / `.env.sample` | Committed intentionally as a template | Variable names + example values |
| `docker-compose.yml` / `docker-compose.override.yml` | Always committed | `environment:` block values |
| `k8s/*.yaml` / `helm/values.yaml` / `helm/*/values.yaml` | Always committed | `env:` and `configMap` references |
| `.github/workflows/*.yml` | Always committed | `env:` sections in jobs/steps |
| `fly.toml` | Always committed | `[env]` section |
| `render.yaml` | Always committed | `envVars:` block |
| `railway.json` | Always committed | `env` block |
| `Dockerfile` | Always committed | `ENV` instructions |
| `serverless.yml` / `template.yaml` (SAM) | Always committed | `environment:` keys |

> **Critical rule:** NEVER scan `.env` files. They must be `.gitignore`d and we never request them. Only scan committed template/example files.

#### Name-pattern matching:

Substrate uses a regex-based name classifier to identify service URL variables:

```
High-signal patterns (URL vars):
  *_API_URL, *_SERVICE_URL, *_ENDPOINT, *_BASE_URL,
  *_HOST (when value is a URL), *_API_BASE, *_GATEWAY_URL

Medium-signal patterns (may be third-party):
  *_API_KEY, *_TOKEN, *_SECRET → only useful if value domain matches known internal service

Infrastructure patterns (non-service dependencies):
  DATABASE_URL, REDIS_URL, KAFKA_BROKER_URL → dependency on a data store, not a repo
```

#### Value-based URL resolution:

Once we have a value (e.g. `https://users.myorg.com`), we look it up in the **URL→Repo Registry** (populated from Phase 3):

```
URL→Repo Registry:
  api.myorg.com          → myorg/backend-api
  users.myorg.com        → myorg/users-service
  payments.myorg.com     → myorg/payments-service
  users-service:8080     → myorg/users-service  (internal Docker DNS)
  users-service.default.svc.cluster.local → myorg/users-service  (k8s DNS)
```

If the URL in the env var matches a registry entry → dependency confirmed.

#### Populating the URL→Repo Registry:

The bootstrapping problem: how does Substrate know which URLs map to which repos?

**Method 1 — `substrate.yaml` deployment declaration** (add to Phase 3 config schema):
```yaml
# substrate.yaml in users-service repo
service: users-service
deployed_urls:
  - https://users.myorg.com             # public URL
  - http://users-service                # internal Docker Compose name
  - users-service.default.svc.cluster.local  # k8s DNS name
  - http://localhost:8090               # local dev port
```

**Method 2 — Terraform output scraping:**
```hcl
output "api_url" {
  value = aws_api_gateway_stage.users.invoke_url
  # Substrate reads this output → maps to the repo that owns this Terraform
}
```

**Method 3 — GitHub Deployments API:**
GitHub's deployment API records the deployment URL when a CI pipeline deploys a service.
Substrate reads `GET /repos/{owner}/{repo}/deployments` and its `statuses` (which include `environment_url`) to automatically register deployed URLs.

---

### 1B. Package Manifest Analysis

Internal SDKs are the most deterministic signal — if `frontend` imports `@myorg/users-sdk`, that SDK was generated from `users-service`'s schema. That IS a contract dependency.

| File | Language | What to parse |
|---|---|---|
| `package.json` | JS/TS | `dependencies` + `devDependencies` keys matching `@{org}/*` |
| `go.mod` | Go | `require` directives referencing `github.com/{org}/*` |
| `requirements.txt` / `pyproject.toml` | Python | Lines matching org namespace |
| `pom.xml` | Java | `<groupId>` matching org convention |
| `Gemfile` | Ruby | `gem` lines matching org source |
| `Cargo.toml` | Rust | `[dependencies]` matching org path |

**SDK→repo resolution:**
```
npm package @myorg/users-sdk → CI pipeline in myorg/users-service publishes this
go module github.com/myorg/payments-go → lives in myorg/payments-service
```

Substrate maintains an **SDK→Repo table** (populated when a repo's CI publishes a package). If the publishing workflow contains `npm publish @myorg/users-sdk`, Substrate registers `@myorg/users-sdk → myorg/users-service`.

**Confidence contribution:** +40 pts (very high signal, very low noise for internal packages)

---

### 1C. OpenAPI Generator Config Files

When a team uses code generation from an OpenAPI spec, the generator config file explicitly names the source spec URL:

```yaml
# .openapi-generator-config.yaml or openapitools.json
inputSpec: "https://raw.githubusercontent.com/myorg/backend-api/main/openapi.yaml"
generatorName: typescript-axios
```

```json
// openapitools.json
{
  "generator-cli": {
    "generators": {
      "users-client": {
        "inputSpec": "https://raw.githubusercontent.com/myorg/users-service/main/openapi.yaml"
      }
    }
  }
}
```

This is the most deterministic possible signal — it is a direct, explicit reference to the provider's schema file.

**Files to scan:**
- `.openapi-generator-config.yaml`
- `openapitools.json`
- `swagger-codegen-config.json`
- `Makefile` targets containing `openapi-generator` or `swagger-codegen` commands
- `package.json` scripts containing generator invocations

**Confidence contribution:** +50 pts (the highest possible single signal — fully deterministic)

---

### 1D. Docker Compose / Kubernetes / Helm

#### Docker Compose
```yaml
services:
  frontend:
    environment:
      - USERS_API_URL=http://users-service:8080  # internal DNS = service dependency
    depends_on:
      - users-service                             # explicit dependency declaration
```

Scan `depends_on` blocks → instant dependency. Scan `environment:` URL values → resolve via URL→repo registry.

**Files:** `docker-compose.yml`, `docker-compose.*.yml`, `compose.yml`

#### Kubernetes
```yaml
# deployment.yaml
env:
  - name: USERS_API_URL
    value: "http://users-service.default.svc.cluster.local"
    # Pattern: {service-name}.{namespace}.svc.cluster.local → repo lookup
```

K8s DNS follows a strict format: `{service-name}.{namespace}.svc.cluster.local`. We can extract `service-name` and look it up in the Kubernetes `Service` manifest registry.

**Files:** `k8s/**/*.yaml`, `helm/**/*.yaml`, `helm/values*.yaml`

**Confidence contribution:** +20 pts each signal (env var match), +30 pts for explicit `depends_on`

---

### 1E. CI/CD Pipeline Analysis

GitHub Actions workflows often wire services together explicitly:

```yaml
# .github/workflows/deploy.yml
jobs:
  deploy:
    env:
      USERS_API_URL: ${{ secrets.USERS_API_URL }}   # variable NAME is a signal
      PAYMENTS_URL: ${{ vars.PAYMENTS_SERVICE_URL }} # variable NAME is a signal
    steps:
      - uses: myorg/users-service-action@v2          # cross-repo action = dependency
```

Even when the value is a secret, the **variable name** pattern (`USERS_API_URL`) is a medium-confidence signal.

Cross-repo `uses:` references (`myorg/other-repo/.github/workflows/shared.yml`) are direct repo-to-repo dependencies.

**Confidence contribution:** +15 pts for name pattern match, +25 pts for cross-repo `uses:` reference

---

## Tier 2 — Infrastructure Signals (Requires Config File Access)

### 2A. Terraform — Extended IaC Analysis

Beyond the `terraform_remote_state` we already capture, Substrate should also scan:

**Environment variable injection in resource definitions:**
```hcl
# AWS ECS task definition
resource "aws_ecs_task_definition" "frontend" {
  container_definitions = jsonencode([{
    environment = [
      { name = "USERS_API_URL", value = aws_api_gateway_stage.users.invoke_url }
      #                                 ↑ references another resource → dependency!
    ]
  }])
}
```

Substrate parses the Terraform expression `aws_api_gateway_stage.users.invoke_url` and resolves `aws_api_gateway_stage.users` to the repo that manages that resource.

**Remote state data sources:**
```hcl
data "terraform_remote_state" "backend_api" {
  backend = "s3"
  config  = { key = "backend-api/terraform.tfstate" }
  # key path "backend-api" → maps to repo myorg/backend-api
}
```

**API Gateway routing rules:**
```hcl
resource "aws_api_gateway_integration" "proxy" {
  uri = "http://users-service.internal:8080/{proxy}"
  # users-service → known repo in registry
}
```

**Confidence contribution:** +25 pts per terraform signal

---

### 2B. Message Queue / Event-Driven Architecture

This is the biggest gap in current tooling — nobody handles event-driven dependencies well.

**Kafka Topic Analysis:**
```hcl
# Terraform MSK or Confluent Cloud config
resource "confluent_kafka_topic" "order_events" {
  topic_name = "orders.v2.order_created"
}
# Producers declare: which service publishes to this topic?
# Consumers declare: which service subscribes to this topic?
```

Substrate maps:
- `orders-service` publishes `orders.v2.*` → any Avro schema changes break consumers
- `payments-service` subscribes to `orders.v2.order_created` → dependency

**AsyncAPI spec cross-references:**
```yaml
# payments-service/asyncapi.yaml
channels:
  orders.v2.order_created:
    subscribe:     # payments-service SUBSCRIBES = dependency on orders-service
      message:
        $ref: 'https://github.com/myorg/orders-service/blob/main/schemas/order_created.avsc'
```

The `$ref` URL is a direct, deterministic dependency link.

**Confidence contribution:** +35 pts for AsyncAPI `$ref`, +20 pts for Kafka topic subscription in Terraform

---

## Tier 3 — Runtime Signals (Post-Deployment Confirmation)

These methods can only discover dependencies after deployment, but they provide the highest-quality confirmation of dependencies already found statically.

### 3A. OpenTelemetry / Distributed Tracing

- Ingest traces from OTel Collector, Datadog APM, Honeycomb, Jaeger
- Extract service-to-service HTTP spans: `service.name: frontend` → `http.url: https://users.myorg.com`
- Cross-reference `http.url` against URL→Repo registry
- **Promotes** an existing dependency from "unconfirmed" to "high confidence"

**Confidence contribution:** +20 pts (confirmation signal, not discovery signal)

### 3B. eBPF / Service Mesh

- Istio/Envoy access logs: `source_workload: frontend` → `destination_service: users-service`
- Very high accuracy but requires k8s + privileged access
- Enterprise-only feature

**Confidence contribution:** +20 pts (confirmation signal)

### 3C. GitHub Deployments API

When a CI pipeline deploys a service, GitHub records the deployment URL. Substrate reads:
```
GET /repos/myorg/users-service/deployments → status[environment_url] = "https://users.myorg.com"
```
This automatically populates the URL→Repo registry without any manual `deployed_urls` declaration.

---

## Implementation Architecture

```
┌─────────────────────────────────────────────────────────┐
│                  Discovery Engine                        │
│                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │  Env Var     │  │  Package     │  │  OpenAPI     │  │
│  │  Scanner     │  │  Manifest    │  │  Generator   │  │
│  │              │  │  Scanner     │  │  Scanner     │  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  │
│         │                 │                  │           │
│  ┌──────┴─────────────────┴──────────────────┴───────┐  │
│  │              Signal Aggregator                     │  │
│  │  (deduplicates edges, sums confidence scores)      │  │
│  └──────────────────────┬────────────────────────────┘  │
│                         │                               │
│  ┌──────────────────────▼────────────────────────────┐  │
│  │              URL → Repo Resolver                   │  │
│  │  (cross-references extracted URLs against registry)│  │
│  └──────────────────────┬────────────────────────────┘  │
│                         │                               │
│  ┌──────────────────────▼────────────────────────────┐  │
│  │         Dependency Graph Writer                    │  │
│  │  (writes to Phase 3 Registry: dependencies table)  │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### URL→Repo Registry Population Sources

```
┌──────────────────────────────────────────────────────────┐
│              URL→Repo Registry                           │
│                                                          │
│  Populated by:                                           │
│  1. substrate.yaml deployed_urls declaration             │
│  2. GitHub Deployments API (automatic, no config needed) │
│  3. Terraform output scraping (invoke_url outputs)       │
│  4. fly.toml app name → *.fly.dev URL                    │
│  5. render.yaml service name → *.onrender.com URL        │
│                                                          │
│  Format:                                                 │
│  "https://users.myorg.com" → myorg/users-service         │
│  "http://users-service:8080" → myorg/users-service       │
│  "users-service.default.svc.cluster.local" → same        │
└──────────────────────────────────────────────────────────┘
```

---

## Confidence Score Reference

| Signal | Points | Rationale |
|---|---|---|
| OpenAPI generator `inputSpec` URL | +50 | 100% deterministic — direct spec reference |
| Internal package import (`@myorg/*`) | +40 | Very strong — SDK = contract |
| `.env.example` URL value resolves in registry | +30 | Value confirmed against known URL |
| `docker-compose depends_on` | +30 | Explicit dependency declaration |
| `terraform_remote_state` key path | +25 | Terraform-declared dependency |
| Terraform env var injection from resource ref | +25 | Terraform-declared wiring |
| AsyncAPI `$ref` cross-repo URL | +35 | Direct schema cross-reference |
| K8s DNS pattern match in env var | +20 | k8s service naming is deterministic |
| `.env.example` variable NAME pattern match | +15 | Name only, value not confirmed |
| CI/CD cross-repo `uses:` | +25 | Direct repo dependency in Actions |
| GitHub Actions env var name pattern | +15 | Name only, value is a secret |
| OTel/Datadog trace confirmation | +20 | Runtime confirmation of static link |
| eBPF/Service mesh confirmation | +20 | Runtime confirmation |

---

## Privacy & Security Rules

1. **Never read `.env` files** — only `.env.example`, `.env.template`, `.env.sample`
2. **Never store env var values** in the registry — only store the resolved dependency edge
3. **Never store secret values** — if a variable name contains `SECRET`, `KEY`, `TOKEN`, `PASSWORD`, skip value extraction
4. **URL→Repo registry** stores only the mapping, never the actual secret credentials
5. All scanning runs inside the GitHub App's existing token scope — no additional permissions required for Tier 1 scanners

---

## Discovery Scanner UI

While the discovery engine runs automatically on the backend via cron and webhook triggers, users must be able to manually trigger scans and view progress for specific repositories.

- **Route:** `/org/[org]/discovery/[repo]`
- **Key Components:**
  - **Scan Status Indicator:** Shows whether a scan is currently active, failed, or completed.
  - **"Run Full Scan" Button:** Manually triggers a comprehensive scan (Env Vars, Packages, Terraform, Kafka, OTel) against the target repository.
  - **Results Panel:** Displays newly discovered edges and their confidence scores immediately after a scan completes.

---

## Rollout Plan

| Phase | What ships | Unblocked by |
|---|---|---|
| **5a** | Env var scanner (`.env.example`, Docker Compose, K8s) + URL→Repo registry (GitHub Deployments API + `substrate.yaml declared_urls`) | Phase 3 registry live |
| **5b** | Package manifest scanner + OpenAPI generator config scanner | 5a |
| **5c** | Terraform extended analysis (env injection + outputs) + Confidence scoring UI in dashboard | 5b |
| **5d** | Kafka/AsyncAPI topic mapping + message queue discovery | 5c |
| **5e** | OTel/Datadog ingestion (runtime confirmation) + eBPF (enterprise) | 5d |

> **Key principle:** Each tier independently improves the graph. Tier 1 (static, in-repo) ships first and delivers value immediately with zero infrastructure requirements. Tiers 2–3 layer on top for progressively higher confidence scores.
