# Contributing to Substrate

First off, thank you for considering contributing to Substrate! 

As a tool designed to enforce Data Contracts and dependency intelligence, we strongly believe in practicing what we preach. Therefore, developing Substrate itself comes with a very strict engineering gate.

## 🚧 The Golden Rule: 100% Spec-First Approach

When developing features for Substrate, we mandate a **100% Spec-First Approach**.

No code changes should be written until the corresponding specification has been designed, discussed, and validated.

### How this works in practice:

1. **Design the Spec First:** If you are adding a new API endpoint, updating a database table, or modifying an event payload, your first commit MUST be to the specification files (e.g., `openapi.yaml`, GraphQL schema, SQL migration files, or internal interface definitions).
2. **Review the Spec:** The specification change is reviewed as a contract. Does it break backward compatibility? Does it handle edge cases? 
3. **Implement the Code:** Only AFTER the specification is validated and approved should you begin writing the application code (Go/TypeScript) that implements or consumes the spec.
4. **CI/CD Validation:** Our CI pipelines will automatically fail if they detect code changes that alter the implicit schema without a corresponding explicit spec change.

By following this, we ensure that Substrate's own architecture remains as reliable and traceable as the systems it protects.

### Pull Request Process

1. Ensure your PR description clearly links to the specification you are implementing.
2. If your PR changes a spec, tag it with `[SPEC]` in the title.
3. Ensure all tests pass.
