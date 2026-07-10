# Substrate — SQL Breaking Change Rules Specification (P1b-T01)

> **Status:** APPROVED ✅
> **Spec-First Gate:** Jules MUST NOT implement P1b-T02 or P1b-T03 until this document is approved.
> **Scope:** PostgreSQL DDL schema snapshot diffing. Detects breaking changes between two `schema.sql` files.
> **Depends on:** `docs/specs/diff-report-schema.md` (output contract), `docs/specs/breaking-change-rules.md` (severity definitions)

---

## Problem Statement

Teams using PostgreSQL routinely break downstream services by:
- Dropping or renaming columns that ORMs or raw queries reference
- Adding `NOT NULL` columns without a `DEFAULT`, causing backend INSERTs to fail
- Adding new `UNIQUE`/`CHECK` constraints that invalidate existing data

These changes are invisible in OpenAPI diffs because they happen at the database layer. Substrate needs a SQL diff engine that catches them in CI before they reach production.

---

## Input Format

Substrate SQL diffing accepts **two PostgreSQL DDL schema snapshot files** — plain `.sql` files produced by:

```bash
pg_dump --schema-only --no-owner --no-privileges -d mydb > schema.sql
```

Both Liquibase and Go ORM teams commit this file to their repository. The CI action fetches the base branch version automatically (same `git show origin/$GITHUB_BASE_REF:schema.sql` pattern already used for OpenAPI).

**MVP Scope (Phase 1b):** PostgreSQL dialect only.
**Future (Phase 2+):** Auto-generate via Docker Postgres container (Option B).

---

## CLI Interface

No changes to the CLI signature. The `schema_type` field in `substrate.yaml` controls which parser is used:

```yaml
# substrate.yaml
service: billing-service
schema_type: sql        # ← triggers the SQL parser
spec_path: schema.sql
```

Command invocation:
```bash
substrate diff schema_base.sql schema_head.sql --format text
```

The engine auto-detects `schema_type: sql` from the config, or falls back to file extension detection (`.sql` → SQL parser).

---

## Internal Architecture

```
schema.sql (base)          schema.sql (head)
      │                           │
      └──────── SQL Parser ───────┘
                    │
              SQLSchema IR
            (Go struct tree)
                    │
            Diff Comparator
                    │
           Raw Change List
                    │
            SQL Rule Engine
                    │
              DiffReport
```

### Parser Library

Use `github.com/pganalyze/pg_query_go/v6` — the official PostgreSQL parser (same C parser as PostgreSQL itself) wrapped for Go. It handles all DDL statements: `CREATE TABLE`, `ALTER TABLE`, `CREATE INDEX`, `CREATE VIEW`, `CREATE TYPE`, etc.

---

## Internal Representation (IR)

Parse the DDL into this Go struct tree in `engine/internal/sql/schema.go`:

```go
package sql

type SQLSchema struct {
    Tables    map[string]*Table
    Views     map[string]*View
    Enums     map[string]*EnumType   // PostgreSQL custom ENUM types
    Sequences map[string]*Sequence
}

type Table struct {
    Name        string
    Columns     map[string]*Column
    Constraints map[string]*Constraint  // PK, UNIQUE, CHECK, FK
    Indexes     map[string]*Index
}

type Column struct {
    Name         string
    DataType     string   // normalized: "varchar", "integer", "timestamp", etc.
    Length       *int     // VARCHAR(n) → n; nil if not applicable
    Precision    *int     // NUMERIC(p,s) → p
    Scale        *int     // NUMERIC(p,s) → s
    NotNull      bool
    Default      *string  // nil = no default
    IsPrimaryKey bool
}

type Constraint struct {
    Name       string
    Type       string   // "PRIMARY_KEY", "UNIQUE", "CHECK", "FOREIGN_KEY"
    Columns    []string
    CheckExpr  *string  // for CHECK constraints
    RefTable   *string  // for FOREIGN KEY
    RefColumns []string
}

type Index struct {
    Name    string
    Columns []string
    Unique  bool
    Where   *string // partial index condition
}

type View struct {
    Name       string
    Definition string   // raw SQL query
    Columns    []string // derived from definition
    Materialized bool
}

type EnumType struct {
    Name   string
    Values []string
}

type Sequence struct {
    Name      string
    DataType  string
    Start     int64
    Increment int64
}
```

**Data type normalization:** All type names are lowercased and aliases resolved:
- `INT` → `integer`
- `INT4` → `integer`
- `INT8` → `bigint`
- `BOOL` → `boolean`
- `TEXT` → `text`
- `CHARACTER VARYING(n)` → `varchar` (length: n)

---

## Breaking Change Rules

### Severity Definitions (inherited from breaking-change-rules.md)

| Severity | CI Action |
|---|---|
| `BREAKING` | Block merge (exit code 2) |
| `WARNING` | Warn, allow merge (exit code 1) |
| `SAFE` | Silent pass (exit code 0) |

---

### 1. Table Rules

#### TABLE_REMOVED
| | |
|---|---|
| **Rule ID** | `TABLE_REMOVED` |
| **Severity** | `BREAKING` |
| **Pattern** | A table present in base is absent in head |
| **Rationale** | Every query, ORM model, or view referencing this table will crash. |

**Example:**
```sql
-- Base
CREATE TABLE orders (id BIGINT PRIMARY KEY, total NUMERIC);

-- Head: orders table is gone
```

---

#### TABLE_ADDED
| | |
|---|---|
| **Rule ID** | `TABLE_ADDED` |
| **Severity** | `SAFE` |
| **Pattern** | A table is in head but not in base |
| **Rationale** | Additive. No existing code references a table that didn't exist. |

---

#### TABLE_RENAMED
| | |
|---|---|
| **Rule ID** | `TABLE_RENAMED` |
| **Severity** | `BREAKING` |
| **Pattern** | A table disappears and a new table with similar name (Levenshtein similarity > 0.7) appears with the same column set |
| **Rationale** | Treat as `TABLE_REMOVED` + `TABLE_ADDED`. All existing queries use the old name. |
| **Note** | If confidence < 0.7, emit `TABLE_REMOVED` + `TABLE_ADDED` separately. |

---

### 2. Column Rules

#### COLUMN_REMOVED
| | |
|---|---|
| **Rule ID** | `COLUMN_REMOVED` |
| **Severity** | `BREAKING` |
| **Pattern** | A column present in base is absent in head |
| **Rationale** | ORM models mapping this column will throw. Raw `SELECT col` queries will fail with "column does not exist". |

**Example:**
```sql
-- Base
CREATE TABLE users (id BIGINT, email TEXT, phone TEXT);

-- Head
CREATE TABLE users (id BIGINT, email TEXT);
-- phone column removed → BREAKING
```

---

#### COLUMN_ADDED_NULLABLE
| | |
|---|---|
| **Rule ID** | `COLUMN_ADDED_NULLABLE` |
| **Severity** | `SAFE` |
| **Pattern** | A new column is added with `NULL` allowed OR with a `DEFAULT` value (making it safe for existing rows) |
| **Rationale** | Existing `INSERT` statements that don't mention this column will succeed. Existing rows get `NULL` or the default. |

---

#### COLUMN_ADDED_NOT_NULL_NO_DEFAULT
| | |
|---|---|
| **Rule ID** | `COLUMN_ADDED_NOT_NULL_NO_DEFAULT` |
| **Severity** | `BREAKING` |
| **Pattern** | A new `NOT NULL` column is added without a `DEFAULT` value |
| **Rationale** | Any `INSERT` from old backend code that doesn't include this column will be rejected by PostgreSQL. This is one of the most common production incidents. |

**Example:**
```sql
-- Head adds:
ALTER TABLE orders ADD COLUMN customer_id BIGINT NOT NULL;
-- No DEFAULT → BREAKING: old INSERTs will fail
```

**Safe version:**
```sql
ALTER TABLE orders ADD COLUMN customer_id BIGINT NOT NULL DEFAULT 0;
-- Has DEFAULT → SAFE
```

---

#### COLUMN_RENAMED
| | |
|---|---|
| **Rule ID** | `COLUMN_RENAMED` |
| **Severity** | `BREAKING` |
| **Pattern** | A column disappears and a column with similar name (similarity > 0.7) appears in the same table |
| **Rationale** | All ORM field mappings and raw queries using the old name break. |

---

#### COLUMN_TYPE_CHANGED
| | |
|---|---|
| **Rule ID** | `COLUMN_TYPE_CHANGED` |
| **Severity** | `BREAKING` |
| **Pattern** | A column's data type changes in a non-widening way (e.g., `VARCHAR` → `INTEGER`, `TEXT` → `BOOLEAN`, `TIMESTAMP` → `DATE`) |
| **Rationale** | ORM deserialization fails. Existing data may be incompatible with the new type. |

**Implementation Note (`pg_query_go` AST Pointers):**
Because `pg_query_go` allocates new memory pointers for AST components like `*int` for `Length` or `Precision` during parsing, diffing logic MUST safely dereference values (`*bCol.Length != *hCol.Length`) instead of directly comparing memory addresses (`bCol.Length != hCol.Length`). Comparing pointer addresses will result in catastrophic false positive `COLUMN_TYPE_CHANGED` evaluations for identical types like `VARCHAR(255)`.

---

#### COLUMN_TYPE_WIDENED
| | |
|---|---|
| **Rule ID** | `COLUMN_TYPE_WIDENED` |
| **Severity** | `WARNING` |
| **Pattern** | A column's data type widens (e.g., `INTEGER` → `BIGINT`, `VARCHAR(50)` → `VARCHAR(200)`, `NUMERIC(10,2)` → `NUMERIC(18,4)`) |
| **Rationale** | PostgreSQL handles widening safely in most cases, but ORMs with strict type mappings may behave unexpectedly. Requires human review. |

**Widening pairs (SAFE direction only):**
- `smallint` → `integer` → `bigint`
- `real` → `double precision`
- `varchar(n)` → `varchar(m)` where `m > n`
- `numeric(p,s)` → `numeric(p2,s)` where `p2 > p`

---

#### COLUMN_MADE_NOT_NULL
| | |
|---|---|
| **Rule ID** | `COLUMN_MADE_NOT_NULL` |
| **Severity** | `BREAKING` |
| **Pattern** | A previously nullable column gains `NOT NULL` constraint |
| **Rationale** | Any existing backend code that inserts `NULL` into this column will be rejected. Existing `NULL` rows may also block the migration itself in PostgreSQL (requires `UPDATE ... SET col = default WHERE col IS NULL` first). |

---

#### COLUMN_DEFAULT_REMOVED
| | |
|---|---|
| **Rule ID** | `COLUMN_DEFAULT_REMOVED` |
| **Severity** | `BREAKING` |
| **Pattern** | `DEFAULT` is removed from a `NOT NULL` column |
| **Rationale** | Any `INSERT` that previously relied on the default will now fail because the column has no fallback value. |

---

#### COLUMN_DEFAULT_CHANGED
| | |
|---|---|
| **Rule ID** | `COLUMN_DEFAULT_CHANGED` |
| **Severity** | `WARNING` |
| **Pattern** | The `DEFAULT` value of a column changes |
| **Rationale** | Backend code relying on the old default may produce different data silently. Not immediately breaking but worth reviewing. |

---

#### COLUMN_NULLABLE_CHANGED
| | |
|---|---|
| **Rule ID** | `COLUMN_NULLABLE_CHANGED` |
| **Severity** | `SAFE` |
| **Pattern** | A `NOT NULL` column becomes nullable |
| **Rationale** | More permissive. Existing inserts still work. Code that previously relied on the guarantee of non-null may need updating, but no crash. |

---

### 3. Constraint Rules

#### PRIMARY_KEY_CHANGED
| | |
|---|---|
| **Rule ID** | `PRIMARY_KEY_CHANGED` |
| **Severity** | `BREAKING` |
| **Pattern** | The set of columns in a table's `PRIMARY KEY` changes |
| **Rationale** | ORM entity identity is tied to the PK. Foreign key references in other tables become invalid. |

---

#### CONSTRAINT_ADDED_UNIQUE
| | |
|---|---|
| **Rule ID** | `CONSTRAINT_ADDED_UNIQUE` |
| **Severity** | `BREAKING` |
| **Pattern** | A new `UNIQUE` constraint is added to a table |
| **Rationale** | Any existing duplicate data will cause the migration itself to fail. After migration, new writes that create duplicates will be rejected. |

---

#### CONSTRAINT_ADDED_CHECK
| | |
|---|---|
| **Rule ID** | `CONSTRAINT_ADDED_CHECK` |
| **Severity** | `BREAKING` |
| **Pattern** | A new `CHECK` constraint is added |
| **Rationale** | Any backend write that doesn't satisfy the check expression will be rejected. |

---

#### FOREIGN_KEY_ADDED
| | |
|---|---|
| **Rule ID** | `FOREIGN_KEY_ADDED` |
| **Severity** | `BREAKING` |
| **Pattern** | A new `FOREIGN KEY` constraint is added |
| **Rationale** | Any insert or update that references a non-existent row in the parent table will be rejected. |

---

#### CONSTRAINT_REMOVED
| | |
|---|---|
| **Rule ID** | `CONSTRAINT_REMOVED` |
| **Severity** | `WARNING` |
| **Pattern** | A `UNIQUE`, `CHECK`, or `FOREIGN KEY` constraint is dropped |
| **Rationale** | Weakens data integrity guarantees. Not immediately breaking for existing code but signals a policy change worth reviewing. |

---

### 4. Index Rules

#### INDEX_REMOVED
| | |
|---|---|
| **Rule ID** | `INDEX_REMOVED` |
| **Severity** | `WARNING` |
| **Pattern** | An index is dropped |
| **Rationale** | Queries relying on this index for performance will slow down significantly. Not a contract break but a meaningful production concern. |

---

#### INDEX_ADDED
| | |
|---|---|
| **Rule ID** | `INDEX_ADDED` |
| **Severity** | `SAFE` |
| **Pattern** | A new index is added |
| **Rationale** | Pure performance improvement. No existing behavior changes. |

---

### 5. View Rules

#### VIEW_REMOVED
| | |
|---|---|
| **Rule ID** | `VIEW_REMOVED` |
| **Severity** | `BREAKING` |
| **Pattern** | A view is dropped |
| **Rationale** | Any query or ORM model targeting this view will fail with "relation does not exist". |

---

#### VIEW_COLUMN_REMOVED
| | |
|---|---|
| **Rule ID** | `VIEW_COLUMN_REMOVED` |
| **Severity** | `BREAKING` |
| **Pattern** | The view's definition changes such that a column it previously exposed no longer exists |
| **Rationale** | Code reading that column from the view gets a runtime error. |

---

#### VIEW_ADDED
| | |
|---|---|
| **Rule ID** | `VIEW_ADDED` |
| **Severity** | `SAFE` |
| **Pattern** | A new view is created |
| **Rationale** | Additive. No existing code references it. |

---

### 6. PostgreSQL ENUM Rules

#### ENUM_VALUE_REMOVED
| | |
|---|---|
| **Rule ID** | `SQL_ENUM_VALUE_REMOVED` |
| **Severity** | `BREAKING` |
| **Pattern** | A value is removed from a PostgreSQL `CREATE TYPE ... AS ENUM` |
| **Rationale** | Existing rows storing that value become invalid. Backend code with exhaustive switch statements break. |

**Note:** Rule ID uses `SQL_` prefix to distinguish from OpenAPI `ENUM_VALUE_REMOVED`.

---

#### ENUM_VALUE_ADDED
| | |
|---|---|
| **Rule ID** | `SQL_ENUM_VALUE_ADDED` |
| **Severity** | `WARNING` |
| **Pattern** | A new value is added to a PostgreSQL ENUM type |
| **Rationale** | ORM enum mappings that are exhaustive (e.g., Go `iota` or TypeScript union types generated from DB schema) will have an unhandled case. |

---

#### ENUM_TYPE_REMOVED
| | |
|---|---|
| **Rule ID** | `SQL_ENUM_TYPE_REMOVED` |
| **Severity** | `BREAKING` |
| **Pattern** | A custom PostgreSQL ENUM type is dropped |
| **Rationale** | Any column using this type is implicitly dropped or invalid. |

---

## Rule ID Quick Reference

| Rule ID | Category | Severity |
|---|---|---|
| `TABLE_REMOVED` | Table | `BREAKING` |
| `TABLE_ADDED` | Table | `SAFE` |
| `TABLE_RENAMED` | Table | `BREAKING` |
| `COLUMN_REMOVED` | Column | `BREAKING` |
| `COLUMN_ADDED_NULLABLE` | Column | `SAFE` |
| `COLUMN_ADDED_NOT_NULL_NO_DEFAULT` | Column | `BREAKING` |
| `COLUMN_RENAMED` | Column | `BREAKING` |
| `COLUMN_TYPE_CHANGED` | Column | `BREAKING` |
| `COLUMN_TYPE_WIDENED` | Column | `WARNING` |
| `COLUMN_MADE_NOT_NULL` | Column | `BREAKING` |
| `COLUMN_DEFAULT_REMOVED` | Column | `BREAKING` |
| `COLUMN_DEFAULT_CHANGED` | Column | `WARNING` |
| `COLUMN_NULLABLE_CHANGED` | Column | `SAFE` |
| `PRIMARY_KEY_CHANGED` | Constraint | `BREAKING` |
| `CONSTRAINT_ADDED_UNIQUE` | Constraint | `BREAKING` |
| `CONSTRAINT_ADDED_CHECK` | Constraint | `BREAKING` |
| `FOREIGN_KEY_ADDED` | Constraint | `BREAKING` |
| `CONSTRAINT_REMOVED` | Constraint | `WARNING` |
| `INDEX_REMOVED` | Index | `WARNING` |
| `INDEX_ADDED` | Index | `SAFE` |
| `VIEW_REMOVED` | View | `BREAKING` |
| `VIEW_COLUMN_REMOVED` | View | `BREAKING` |
| `VIEW_ADDED` | View | `SAFE` |
| `SQL_ENUM_VALUE_REMOVED` | Enum | `BREAKING` |
| `SQL_ENUM_VALUE_ADDED` | Enum | `WARNING` |
| `SQL_ENUM_TYPE_REMOVED` | Enum | `BREAKING` |

**Total: 26 rules — 14 BREAKING, 6 WARNING, 6 SAFE**

---

## Files To Create / Modify

| File | Change |
|---|---|
| `engine/internal/sql/parser.go` | New — parses DDL text into `SQLSchema` IR using `pg_query_go` |
| `engine/internal/sql/schema.go` | New — IR struct definitions |
| `engine/internal/sql/diff.go` | New — diffs two `SQLSchema` IRs, returns `[]RawChange` |
| `engine/internal/sql/rules.go` | New — maps `RawChange` → `report.Change` applying the 26 rules |
| `engine/internal/sql/parser_test.go` | New — unit tests for parser (valid DDL, edge cases) |
| `engine/internal/sql/diff_test.go` | New — unit tests for all 26 rules with fixture `.sql` pairs |
| `engine/internal/diff/openapi.go` | No change — SQL engine is a parallel package |
| `engine/cmd/substrate/main.go` | Modify — wire SQL parser when `schema_type: sql` or `.sql` extension |
| `engine/go.mod` | Add `github.com/pganalyze/pg_query_go/v6` |

---

## Test Requirements

All tests in `engine/internal/sql/diff_test.go` MUST use fixture `.sql` pairs in `engine/internal/sql/testdata/`:

| Test Case | Base File | Head File | Expected Rule |
|---|---|---|---|
| Table removed | `base_table_removed.sql` | `rev_table_removed.sql` | `TABLE_REMOVED` → BREAKING |
| Column removed | `base_col_removed.sql` | `rev_col_removed.sql` | `COLUMN_REMOVED` → BREAKING |
| Column nullable added | `base_col_add_null.sql` | `rev_col_add_null.sql` | `COLUMN_ADDED_NULLABLE` → SAFE |
| Column NOT NULL no default | `base_col_add_nn.sql` | `rev_col_add_nn.sql` | `COLUMN_ADDED_NOT_NULL_NO_DEFAULT` → BREAKING |
| Column type changed | `base_col_type.sql` | `rev_col_type.sql` | `COLUMN_TYPE_CHANGED` → BREAKING |
| Column type widened | `base_col_widen.sql` | `rev_col_widen.sql` | `COLUMN_TYPE_WIDENED` → WARNING |
| Column made NOT NULL | `base_col_nn.sql` | `rev_col_nn.sql` | `COLUMN_MADE_NOT_NULL` → BREAKING |
| Unique constraint added | `base_unique.sql` | `rev_unique.sql` | `CONSTRAINT_ADDED_UNIQUE` → BREAKING |
| View removed | `base_view.sql` | `rev_view.sql` | `VIEW_REMOVED` → BREAKING |
| Enum value removed | `base_enum.sql` | `rev_enum.sql` | `SQL_ENUM_VALUE_REMOVED` → BREAKING |
| No changes | `base_no_change.sql` | `base_no_change.sql` | 0 changes |
| Invalid SQL | `base_invalid.sql` | any | error, no DiffReport |

---

## How Teams Generate `schema.sql`

### Liquibase Teams
```bash
# Apply changesets to a test DB and dump
liquibase update --url=jdbc:postgresql://localhost/testdb
pg_dump --schema-only --no-owner --no-privileges testdb > schema.sql
git add schema.sql
git commit -m "chore: update schema snapshot"
```

### Go ORM / Raw SQL Teams (GORM, sqlx, raw migrations)
```bash
# Run migrations against a local/CI Postgres instance
go run ./cmd/migrate   # or however you apply migrations
pg_dump --schema-only --no-owner --no-privileges mydb > schema.sql
git add schema.sql
git commit -m "chore: update schema snapshot"
```

### Recommended CI Step (add to your workflow)
```yaml
- name: Validate schema snapshot is up to date
  run: |
    pg_dump --schema-only --no-owner --no-privileges $DATABASE_URL > /tmp/current_schema.sql
    diff schema.sql /tmp/current_schema.sql || (echo "❌ schema.sql is out of date. Run pg_dump and commit." && exit 1)
```

---

## Future: Option B (Phase 2)

> **Not in scope for P1b.** Tracked as a separate task.
>
> Phase 2 will add an optional `substrate-generate-schema` GitHub Action step that:
> 1. Spins up a Docker Postgres container
> 2. Applies all Liquibase changesets OR runs Go ORM auto-migrate
> 3. Runs `pg_dump --schema-only` automatically
> 4. Passes the output directly to `substrate diff`
>
> This eliminates the need for teams to commit `schema.sql` manually.

---

## Next Steps

> Once this spec is approved:
> - Create Jules prompt at `prompts/phase-1b-sql/t01_sql_parser.txt`
> - Submit via `python3 scripts/jules_submit.py --task 101`
> - P1b-T02: SQL DDL parser (`engine/internal/sql/parser.go` + `schema.go`)
> - P1b-T03: SQL rule engine + tests (`engine/internal/sql/diff.go` + `rules.go`)
