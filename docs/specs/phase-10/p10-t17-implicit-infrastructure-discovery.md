# Phase 10 - Task 17: Implicit Infrastructure Discovery (IID)

## 1. Vision & Goal
The ultimate goal of Substrate's enterprise offering is **Zero-Config Architectural Mapping**. Developers should not have to manually document their infrastructure. 

Implicit Infrastructure Discovery (IID) will automatically map a repository's Languages, Frameworks, Databases, Queues, Cloud Services, and external SaaS dependencies simply by analyzing the codebase during the Webhook push event.

## 2. The Four Pillars of Discovery

### Pillar 1: Language & Framework Fingerprinting
By scanning manifest files (`package.json`, `go.mod`, `pom.xml`, `requirements.txt`), Substrate will assign metadata tags to the service node.
- **Languages Supported:** Node.js/TypeScript, Go, Python, Java, Rust.
- **Framework Detection:** e.g., Next.js, Express, Spring Boot, Django, FastAPI, Gin, Axum.
- **UI Benefit:** The graph node will display the language/framework logo (e.g., the Go Gopher or Python logo), giving architects an instant view of the stack diversity.

### Pillar 2: Database & Infrastructure Inference
We will map specific driver libraries to infrastructure nodes:
- **Databases:** PostgreSQL (`pg`, `pq`, `psycopg2`), MongoDB (`mongoose`), Redis (`go-redis`, `ioredis`), Cassandra, ElasticSearch.
- **Event Brokers:** Kafka (`kafkajs`, `confluent-kafka-go`), RabbitMQ (`amqplib`), NATS.
- **Action:** Substrate will automatically create an "Infrastructure Node" on the graph (e.g., `infra:redis`) and draw an edge from the microservice to it.

### Pillar 3: External SaaS & Cloud API Detection
We can detect consumption of 3rd-party APIs by looking for official SDKs:
- **SaaS Platforms:** Stripe (`stripe-node`), Twilio (`twilio`), Slack (`@slack/web-api`), SendGrid.
- **Cloud Primitives:** AWS S3 (`@aws-sdk/client-s3`), GCP PubSub (`@google-cloud/pubsub`).
- **Action:** Substrate will map these to external boundary nodes (e.g., `saas:stripe`).

### Pillar 4: IaC & Container Manifest Correlation (Advanced)
A package might be installed but unused. To achieve 100% accuracy, IID will parse Infrastructure-as-Code files:
- **Docker Compose (`docker-compose.yml`):** Scan for standard images (`postgres:15`, `redis:7`, `confluentinc/cp-kafka`). 
- **Kubernetes Manifests:** Scan Helm charts and K8s YAMLs.
- **Environment Variables:** By statically analyzing the code using Tree-sitter, we can detect variables like `os.Getenv("REDIS_URL")`. If two different repositories look for the same `REDIS_URL` environment variable name, Substrate can infer they are talking to the *same* Redis cluster!

## 3. Execution Pipeline

1. **Webhook Event:** Push event triggers the Cloudflare Worker.
2. **Shallow Clone:** The worker pulls the manifest files (`package.json`, `docker-compose.yml`, etc.) but skips massive directories.
3. **Heuristic Engine:** The Go backend runs the manifest against a pre-compiled mapping dictionary.
4. **Graph Upsert:** The backend upserts the discovered implicit nodes into the Postgres DB alongside the explicitly defined `substrate.yaml` dependencies.
5. **Real-time UI:** The Svelte Flow graph renders the new infrastructure topologies.
