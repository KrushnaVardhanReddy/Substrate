# P15-T10: Schema Insurance (Enterprise Tier Add-On)

## Overview
A premium enterprise add-on: if a breaking change slips through Substrate's monitoring and causes a verified production incident, Substrate pays an SLA credit. This turns Substrate into a risk management instrument.

## Requirements
1. **Insurance Policy Data Model**: Add tables `insurance_policies` and `insurance_claims` to the Postgres database.
2. **Dashboard UI**: Add an "Insurance" settings panel in the Enterprise organization view showing current policy limits and claim history.
3. **Claim Workflow**: Add a button to "File a Claim" linking a GitHub PR (the missed breaking change) to an incident date.
4. **Automated Verification**: The API backend must automatically verify if the PR actually bypassed Substrate checks (e.g., if a developer used `substrate override`). If overridden manually, the claim is rejected automatically.

## API Specifications

**Note:** All paths use the string `{org}` parameter, which represents the organization's unique string name (e.g., `acme`), NOT its UUID. The backend is responsible for looking up the corresponding organization UUID in the database using this name.

### 1. Get Insurance Policy
- **Endpoint**: `GET /api/v1/org/{org}/insurance/policy`
- **Auth**: Requires `authzMW` with a valid JWT possessing the `admin` role for `{org}`.
- **Response** (`200 OK`):
  ```json
  {
    "id": "uuid",
    "org_id": "uuid",
    "tier": "enterprise",
    "limit_cents": 5000000,
    "deductible_cents": 10000,
    "active": true
  }
  ```

### 2. File a Claim
- **Endpoint**: `POST /api/v1/org/{org}/insurance/claims`
- **Auth**: Requires `authzMW` with a valid JWT possessing the `admin` role for `{org}`.
- **Request Body**:
  ```json
  {
    "github_pr_url": "https://github.com/org/repo/pull/42",
    "incident_date": "2026-07-22T00:00:00Z",
    "amount_cents": 50000
  }
  ```
- **Response** (`201 Created`):
  ```json
  {
    "id": "uuid",
    "status": "PENDING",
    ...
  }
  ```
