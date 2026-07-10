# Substrate — Phase 1c: GraphQL Schema Diff Adapter

This specification defines the breaking change rules and architecture for the Substrate GraphQL diff engine. Substrate analyzes `.graphql` or `.gql` SDL (Schema Definition Language) files.

## 1. Architectural Strategy

- **Parser Library:** `github.com/vektah/gqlparser/v2`
- **Why?** It is the underlying MIT-licensed AST engine that powers `gqlgen`. It parses raw SDL strings into a highly traversable `*ast.Schema` object.
- **Adapter Location:** `engine/internal/diff/graphql.go`
- **Output:** Returns a standard `report.DiffReport` with `SchemaType` = `"graphql"`.

---

## 2. GraphQL Breaking Change Rules

Substrate categorizes GraphQL changes into `BREAKING` (breaks downstream consumers, e.g., Apollo Client), `WARNING` (safe but requires attention), and `SAFE` (fully backward compatible).

### 2.1 Types and Interfaces

| Rule ID | Severity | Trigger Condition |
|---------|----------|-------------------|
| `GQL_TYPE_REMOVED` | **BREAKING** | An Object, Input Object, Interface, Union, Enum, or Scalar type was removed. |
| `GQL_FIELD_REMOVED` | **BREAKING** | A field was removed from an Object or Interface type. |
| `GQL_FIELD_TYPE_CHANGED` | **BREAKING** | The return type of a field changed (e.g., `String` to `Int`, `User!` to `User`). Note: Changing from nullable to non-nullable (`User` to `User!`) is SAFE. |
| `GQL_UNION_MEMBER_REMOVED`| **BREAKING** | A type was removed from a Union. |
| `GQL_ENUM_VALUE_REMOVED` | **BREAKING** | A possible value was removed from an Enum type. |
| `GQL_INTERFACE_REMOVED` | **BREAKING** | An object type no longer implements an interface it previously implemented. |

### 2.2 Arguments and Inputs

| Rule ID | Severity | Trigger Condition |
|---------|----------|-------------------|
| `GQL_ARGUMENT_REMOVED` | **BREAKING** | An argument was removed from a field. |
| `GQL_ARGUMENT_TYPE_CHANGED` | **BREAKING** | The type of an argument changed in a backward-incompatible way. |
| `GQL_REQUIRED_ARGUMENT_ADDED`| **BREAKING** | An argument was added to a field that is non-nullable (e.g., `String!`) and has no default value. |
| `GQL_INPUT_FIELD_ADDED_REQUIRED`| **BREAKING** | A non-nullable field without a default value was added to an Input Object type. |
| `GQL_OPTIONAL_ARGUMENT_ADDED` | **SAFE** | An argument was added to a field but it is nullable or has a default value. |
| `GQL_INPUT_FIELD_ADDED_OPTIONAL`| **SAFE** | A nullable field or field with a default value was added to an Input Object type. |

### 2.3 Directives

| Rule ID | Severity | Trigger Condition |
|---------|----------|-------------------|
| `GQL_DIRECTIVE_REMOVED` | **BREAKING** | A custom directive definition was removed. |
| `GQL_DIRECTIVE_LOCATION_REMOVED`| **BREAKING** | A valid location was removed from a custom directive definition. |

### 2.4 Deprecation

| Rule ID | Severity | Trigger Condition |
|---------|----------|-------------------|
| `GQL_FIELD_DEPRECATED` | **WARNING** | A field was marked with `@deprecated`. |
| `GQL_ENUM_VALUE_DEPRECATED` | **WARNING** | An enum value was marked with `@deprecated`. |

---

## 3. Input Validation

The adapter must validate the input schemas using the AST parser before diffing. If `vektah/gqlparser` fails to parse the base or head schema (due to invalid GraphQL syntax), the adapter must return the parser error directly, and the CLI must exit with code `3` (Invalid Schema).
