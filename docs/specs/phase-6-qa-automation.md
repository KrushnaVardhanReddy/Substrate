# Phase 6 Spec: QA & Automation Layer (The SDET Co-Pilot)

> **Status:** 💡 Planned — Phase 6
> **Goal:** Transform Substrate from a CI bug-catcher into a full SDLC Orchestrator that automatically generates and maintains test infrastructure.

---

## 1. Auto-Updating Postman & Cypress Tests

**Problem:** Backend developers rename fields (e.g., `user_id` to `userId`), which safely passes API checks but silently breaks the QA team's Cypress tests and Postman collections overnight.
**Solution:**
- Substrate parses changes to the OpenAPI spec.
- Using the Postman API (or by managing a specific "Auto-Generated" folder in the repo to avoid human merge conflicts), Substrate automatically pushes updates to the Postman collections to reflect the new schema.
- Substrate also allows downloading the generic OpenAPI Spec directly from the QA Dashboard so QA can use alternative tools like Hopscotch, Bruno, or Insomnia in a vendor-neutral way.
- For Cypress, Substrate can auto-generate TypeScript fixture updates and open a PR in the QA repository.

## 2. "Shadow API" Test Coverage

**Problem:** Code coverage tools only show which lines of code ran, not which API fields were actually tested.
**Solution:**
- Substrate ingests Datadog/OTel traffic traces during the E2E test run.
- It maps the actual JSON payloads against the OpenAPI schema.
- The Substrate Dashboard provides a visual "Shadow Coverage" report, color-coding the schema:
  - 🟢 Green: Field tested.
  - 🔴 Red: Field defined in spec but never sent/received during tests.

## 3. Auto-Generating Test Code (Fuzzing)

**Problem:** QA manually writes edge-case payloads to test API constraints.
**Solution:**
- Instead of just generating dummy JSON payloads, Substrate reads constraints (`maxLength`, `minimum`, `required`) and generates executable test code (e.g., Playwright API tests or Go `httpexpect` tests).
- Substrate drops these auto-generated edge-case tests directly into a `.substrate/autotests/` directory in the repository.

## 4. Mock Server Time Machine

**Problem:** Reproducing a bug from an old version of an API is extremely difficult because it requires the entire backend infrastructure and database state to be rolled back.
**Solution:**
- Substrate acts as a "Time Machine" using the versioned Contract Registry.
- A developer runs `substrate mock --env production --timestamp "2026-07-01"`.
- Substrate pulls the exact OpenAPI schema from that timestamp and spins up a local Mock API server instantly.
- Frontend and QA can verify how the client handled the old contract without needing backend infrastructure.
