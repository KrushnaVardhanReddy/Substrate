1. Address feedback in `scripts/e2e/phase16_e2e_test.go`:
   - Replace hardcoded `p16DbURL` with the `TEST_DATABASE_URL` environment variable.
   - Refactor tests to be table-driven (using a struct with Name, URL, Method, Token, ExpectedStatus, etc.).
   - Ensure we don't rely on `createP15JWT` by re-implementing or generalizing it using the existing JWT Secret environment variable logic, or better yet, since the test needs to test `POST /api/v1/sandbox/request` which requires a paseto token from `/api/v1/sandbox/token` (and that one needs a JWT), use standard `createJWT`. Actually I'll implement `createP16JWT` to be safe and independent.
   - Avoid raw SQL inserts for `org` and `repos` by using `helpers.go` `seedOrg` and `seedRepo`. Wait, `helpers.go` doesn't have `seedOrg` / `seedRepo`? Let's check `scripts/e2e/helpers.go` and `scripts/e2e/helpers/` again. Oh wait, the prompt said: `DB seeding for E2E uses the helpers in scripts/e2e/helpers.go (seedOrg, seedRepo).`. I will check if they are in `testdata` or some other package, but I can just define my own `seedOrg` or look in `scripts/e2e/helpers.go`. Let's look again at what's available. If they don't exist, I'll write them, but the prompt says they do. Actually I'll check `scripts/e2e/helpers.go` one more time. In previous searches it only had `SetupP17Database`. I'll run `grep -r seedOrg scripts/e2e` to find it.

2. Address feedback in `scripts/e2e/seed_p16.go`:
   - Use `TEST_DATABASE_URL` instead of `DATABASE_URL` (or default to it).
   - Use helpers if possible, or just raw SQL for `repo_guides` since the prompt said `Seed a repo_guides row directly via SQL in the test setup using db.Exec()`.

3. Re-request code review.
