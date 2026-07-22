# P10-T09: Protobuf & gRPC Schema Registry (Go Engine Parsing)

## Objective
Extend the Substrate Go Engine to natively parse, validate, and store Protocol Buffers (.proto) schemas and gRPC service definitions, bringing first-class support for gRPC-based microservices into the dependency graph.

## Context
Substrate currently excels at mapping HTTP/REST dependencies via OpenAPI and GraphQL schemas. However, many enterprise architectures rely heavily on gRPC for internal service-to-service communication. To provide a complete architectural picture, the core engine must understand `.proto` files and detect breaking changes in gRPC contracts.

## Requirements

### 1. Go Engine Parser
- **AST Parsing:** Implement a parser for `.proto` files (proto2 and proto3 syntax) using a robust Go library (e.g., `github.com/emicklei/proto`).
- **Schema Extraction:** Extract Messages, Services, RPC methods, and Enums into the standard Substrate internal schema representation.
- **Dependency Resolution:** Identify upstream/downstream links based on imported `.proto` files and defined gRPC services.

### 2. Diffing & Impact Analysis
- Implement a diffing algorithm specifically for Protobuf.
- **Breaking Changes:** Correctly identify breaking changes (e.g., changing a field type, removing a field, changing field tags/numbers, removing an RPC method).
- **Non-Breaking Changes:** Identify safe changes (e.g., adding a new optional field, adding a new RPC method).

### 3. Registry & Storage
- Persist the parsed Protobuf schemas in the Substrate database alongside OpenAPI and GraphQL schemas.

## Acceptance Criteria
1. The Go Engine can ingest a `.proto` file and accurately represent its structure in the database.
2. The diffing engine correctly flags removing a field or changing a tag number as a breaking change.
3. The dependency graph correctly maps services that consume the gRPC definitions.
