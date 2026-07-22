# SPEC: CC-T01 — PASETO Security Migration

## Overview
Substrate is migrating its core authentication middleware from JWT (JSON Web Tokens) to PASETO (Platform-Agnostic Security Tokens). As an enterprise API governance tool, relying on JWT exposes the platform to known cryptographic footguns (such as `alg="none"` and cross-algorithm confusion attacks). PASETO solves this by enforcing opinionated, version-locked cryptographic ciphers.

## Scope of Migration
This task is tightly scoped to replacing the internal API authentication mechanism. It sets the foundation for the upcoming Air-Gapped License Validator (P15-T15), which requires PASETO asymmetric signatures (`v4.public`).

### Phase 1: Internal API Auth (`v4.local`)
1. **Dependency Replacement**:
   - Uninstall `github.com/golang-jwt/jwt/v5` from the `api/` package.
   - Install a reputable Go PASETO v4 library (e.g., `github.com/aidarkhanov/nanoid` is strictly for IDs; standard PASETO v4 packages like `github.com/o1egl/paseto` or an equivalent V4-compatible fork should be vetted and used).

2. **Middleware Refactoring (`api/internal/server/middleware.go`)**:
   - `AuthMiddleware`: Currently parses a JWT using the `JWT_SECRET`. It must be updated to parse and validate a PASETO `v4.local` (symmetric encryption) token using the same secret byte array.
   - `AuthzMiddleware`: Must be updated to parse PASETO claims for role-based access control (`admin` vs `member`).

3. **Test Refactoring (`api/internal/server/middleware_test.go`)**:
   - Replace all instances of `jwt.NewWithClaims` with the equivalent PASETO v4 token builder.
   - Ensure expiration time (`exp`) and role claims are properly encoded and decoded.

## Future Context: Phase 15 Licensing
Once this migration is complete, the engine will be ready for **P15-T15**. The upcoming license validator will rely on PASETO `v4.public` tokens (asymmetric Ed25519 signatures) where the licensing server holds the private key, and the local Substrate engine verifies the payload using a compiled-in public key.

## Definition of Done
- `golang-jwt` is completely removed from `go.mod`.
- `api/internal/server/middleware.go` uses PASETO exclusively.
- All HTTP handlers and middleware unit tests pass without relying on JWT generation.
