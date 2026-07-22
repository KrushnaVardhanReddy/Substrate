# P15-T10: Schema Insurance (Enterprise Tier Add-On)

## Overview
A premium enterprise add-on: if a breaking change slips through Substrate's monitoring and causes a verified production incident, Substrate pays an SLA credit. This turns Substrate into a risk management instrument.

## Requirements
1. **Insurance Policy Data Model**: Add tables `insurance_policies` and `insurance_claims` to the Postgres database.
2. **Dashboard UI**: Add an "Insurance" settings panel in the Enterprise organization view showing current policy limits and claim history.
3. **Claim Workflow**: Add a button to "File a Claim" linking a GitHub PR (the missed breaking change) to an incident date.
4. **Automated Verification**: The API backend must automatically verify if the PR actually bypassed Substrate checks (e.g., if a developer used `substrate override`). If overridden manually, the claim is rejected automatically.
