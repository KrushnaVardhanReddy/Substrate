# Architecture & Design

## Hexagonal Architecture
Substrate uses a strictly isolated Hexagonal (Ports and Adapters) architecture.
- **Ports**: Interface definitions live in `api/internal/ports/`.
- **Adapters**: The database layer is implemented using `sqlc` to generate type-safe queries in `api/internal/db/queries/`.
- HTTP handlers must only communicate through port interfaces, never interacting directly with the DB layer. This allows the monolithic `Store` to be split into domain-specific interfaces (e.g., `RepoStore`, `ChangeStore`).

## Configuration Management
The project uses `viper` as the single source of truth for all configurations (replacing direct `os.Getenv` calls).

## GraphQL Supergraphs & EBPF
Substrate natively diffs Apollo Federation supergraphs. It also supports zero-latency runtime drift detection via `cilium/ebpf` kernel probes.
