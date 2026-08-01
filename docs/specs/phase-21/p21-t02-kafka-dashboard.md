# Phase 21 — Task 02: Kafka Schema Governance Dashboard

## 1. Overview

Svelte 5 dashboard page at `/org/{org}/kafka` showing registered Schema Registry
sources, tracked Kafka subjects, version history, and a live breaking-change event feed.

## 2. Technical Requirements

### 2.1 Page: `/org/{org}/kafka`

Three sections:
1. **Sources** — list of registered registries + last-event timestamp. "Add Registry" modal posts new source and displays the generated webhook URL + token.
2. **Subjects** — paginated table: Subject, Type (AVRO/PROTO/JSON), Latest Version, Last Changed, Breaking Changes count.
3. **Live Breaking Changes Feed** — SSE-connected feed from `kafka_breaking_change` events.

### 2.2 Subject Detail Drawer

Clicking a subject opens a side drawer:
- Version history timeline.
- Side-by-side diff of any two versions (reuse existing `DiffViewer` component).
- Blast radius list of known consumers.

### 2.3 Graph Integration

When a `kafka_breaking_change` SSE event arrives, highlight the affected subject
as a new `kafka_schema` node type in the dependency graph with a distinct icon
(`lucide-svelte/icons/radio`).

## 3. Svelte 5 Rules

- All state via `$state`, `$derived`, `$effect` runes.
- `import Radio from 'lucide-svelte/icons/radio'` — direct import only.
- SSE cleanup in `$effect` return function.
- Add Kafka nav item to org sidebar.

## 4. Deliverables

1. `dashboard/src/routes/(app)/org/[org]/kafka/+page.svelte` (CREATE)
2. `dashboard/src/routes/(app)/org/[org]/kafka/+page.ts` (CREATE)
3. `dashboard/src/lib/components/KafkaSourceCard.svelte` (CREATE)
4. `dashboard/src/lib/components/KafkaSubjectTable.svelte` (CREATE)
5. `dashboard/src/lib/components/KafkaBreakingFeed.svelte` (CREATE)
6. MODIFY: org sidebar — add Kafka nav item
7. `dashboard/src/lib/components/KafkaSourceCard.test.ts` (CREATE)
8. Run `npm run check` and `npx vitest run`.
