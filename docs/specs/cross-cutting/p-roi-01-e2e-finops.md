# P-ROI-01: E2E FinOps ROI API

## Goal
Validate the FinOps backend APIs (`GET /api/v1/telemetry/roi/{org}`, `POST /api/v1/finops/predict`).

## Scenarios
1. **ROI Fetch**: `GET` returns the calculated financial savings from blocked breaking changes.
2. **Predict Cost**: `POST` to predict returns an estimated incident cost based on a payload of affected repositories.

## Constraints
- File: `scripts/e2e/phase_roi_e2e_test.go`
- Must use live PostgreSQL database connections, no mocks.
