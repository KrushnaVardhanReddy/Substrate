# Spec: P12-T06 — 1,000-Node UI Stress Test

## 1. Overview
Expand the scale generator to produce a 1,000-node / 3,000-edge deterministic acyclic graph and validate it passes through the Cytoscape canvas without browser lockup.

## 2. Owner Split (No Conflict)
- **Jules** → `scripts/e2e/scale_generator.go` (Go generator + unit test)
- **Stitch** → No changes required. `dashboard/tests/stress-test.spec.ts` already passes with the existing stress-test org.

## 3. Files Modified
- `scripts/e2e/scale_generator.go` ← expand the existing generator

## 4. Jules Requirements

### Scale Generator Update
Update `scale_generator.go` to support a 1,000-node graph:

**Graph Construction Rules:**
1. **Backbone chain:** `node-0 → node-1 → node-2 → … → node-999` (999 edges).
2. **Skip edges:** Add 2,001 additional directed edges `node-i → node-j` where `j > i + 1` (strict forward direction prevents cycles). Use a seeded PRNG (`rand.New(rand.NewSource(42))`) for determinism.
3. **Total:** 1,000 nodes, ~3,000 edges.
4. **Status distribution:** 5% `BREAKING`, 95% `SAFE` (use `i % 20 == 0`).

**Output format:**
```json
[
  { "provider": "node-0", "consumer": "node-1", "status": "SAFE" },
  ...
]
```

### Unit Test `TestScaleGenerator_1000Nodes`
1. Call the generator with `n=1000, edges=3000`.
2. Assert output array length >= 2,000 (backbone + skip edges).
3. Assert no duplicate `provider+consumer` pairs.
4. Assert no cycles via DFS from `node-0`: must terminate without revisiting any node.
5. Assert exactly `1000` unique node IDs appear across all entries.

## 5. Stitch Note (No Action Needed)
The existing `stress-test.spec.ts` hits the `/org/stress-test/graph` route which uses in-code procedural generation (in `+page.ts`). It already passes. No Playwright changes are needed for this task.

## 6. Success Criteria
- `go test ./scripts/e2e/ -run TestScaleGenerator_1000Nodes` passes.
- Generated JSON is valid and cycle-free.
- `npx playwright test tests/stress-test.spec.ts` continues to pass (no regression).
