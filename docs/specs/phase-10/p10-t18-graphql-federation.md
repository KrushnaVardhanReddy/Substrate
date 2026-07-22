# P10-T18: GraphQL Supergraph Federation

## 1. Objective
Extend Substrate's existing GraphQL diffing engine to natively support **Apollo Federation**. 
Enterprise architectures rarely use a single monolithic GraphQL schema; they compose multiple "Subgraphs" (e.g., Accounts, Inventory, Payments) into a single "Supergraph" router. Substrate must detect when a schema change in a single Subgraph will cause the global Supergraph composition to fail or break downstream clients.

## 2. Core Federation Concepts to Support
Substrate's AST parser must correctly interpret Apollo Federation v2 directives:
- `@key(fields: "id")`: Defines the primary key for an entity. Deleting a key field is a **Critical Breaking Change**.
- `@external`: Declares a field owned by another Subgraph. If the providing Subgraph drops this field, it breaks the requesting Subgraph.
- `@requires(fields: "...")`: Declares that this resolver needs data from another Subgraph.
- `@provides(fields: "...")`: Declares that this resolver can provide data for a field usually resolved elsewhere.
- `@shareable`: Indicates multiple Subgraphs can resolve this field. Dropping it from one Subgraph is *not* breaking if another still provides it.

## 3. Architecture & Implementation
1. **Engine AST Updates (`engine/internal/graphql/`):**
   - Update the GraphQL parser (`github.com/vektah/gqlparser/v2`) to load Federation directive definitions before parsing. If they are absent, automatically inject the Apollo Federation v2 `extend schema @link(...)` and directive definitions so the parser doesn't throw syntax errors.
2. **Federation-Aware Diffing Rules:**
   - **Rule F1 (Key Dropped):** If a field marked as a `@key` is removed or its type is changed, flag as `FederationKeyBroken`.
   - **Rule F2 (External Dependency Broken):** If Repo A drops field `price` from `Product`, but Repo B has `extend type Product { price: Float @external }`, Substrate must flag this cross-repo breakage. (Requires querying the Substrate Registry API for consumers holding `@external` references).
3. **CLI UX:**
   - Introduce a new schema type flag: `substrate diff base.graphql head.graphql --schema-type=graphql-federation`.

## 4. Deliverables for Jules
- Modify `engine/internal/graphql/diff.go` to handle Federation directives.
- Modify `api/internal/handlers/impact.go` to support Federation-specific cross-repo checks (Rule F2).
- Add unit tests demonstrating a `@key` removal correctly failing the check suite.
