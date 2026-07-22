# Spec: P12-T04 — SSE Connection Resilience Test

## 1. Overview
Write Go tests to validate that the `SSEBroker` correctly handles 100+ concurrent client connections and abrupt disconnects without memory leaks or goroutine deadlocks.

## 2. Owner
**Jules** (Go Backend)

## 3. Files Modified
- `scripts/e2e/phase12_sse_test.go` ← expand this file only

## 4. Requirements

### Test A — 100 Concurrent Clients Registration
1. Create a new `SSEBroker` via `NewSSEBroker()`.
2. Call `broker.Start()` in a goroutine.
3. Register 100 client channels by sending to `broker.newClients`.
4. Assert `len(broker.clients) == 100` after a short sync.

### Test B — Abrupt Disconnect (No Memory Leak)
1. Register 100 clients as above.
2. For each client, send its channel to `broker.defunctClients` to simulate disconnect.
3. Allow the broker 500ms to process removals.
4. Assert `len(broker.clients) == 0`.
5. Assert no goroutines are leaked (use `goleak` or `runtime.NumGoroutine()` comparison).

### Test C — Broadcast Under Concurrent Disconnects
1. Register 50 clients.
2. In parallel goroutines: disconnect 25 of them AND broadcast a message via `broker.messages`.
3. Assert no deadlock occurs within a 2-second `time.After` timeout.
4. Assert no `panic`.

### Test D — Heartbeat Tick
1. Register 1 client channel with a buffer of at least 2.
2. Wait 6 seconds (heartbeat fires every 5s per spec).
3. Assert the client channel received a message containing `"heartbeat"`.

## 5. Technical Constraints
- All tests MUST use `t.Parallel()` where safe.
- The broker's internal fields (`clients`, `newClients`, `defunctClients`, `messages`) must be accessible for assertions — if they are unexported, expose a `Len() int` method on `SSEBroker`.
- Use `-race` flag: `go test -race -count=1 ./scripts/e2e/...`.
- Do NOT start a real HTTP server. Test the broker struct directly.

## 6. Success Criteria
- `go test -race -count=1 ./scripts/e2e/ -run TestSSE` passes with exit code 0.
- No data races detected under `-race`.
- No goroutine leaks detected in Test B.
