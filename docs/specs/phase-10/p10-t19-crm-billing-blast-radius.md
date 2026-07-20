# P10-T19: CRM/Billing Blast Radius

## Overview
Integrate with Stripe and Salesforce to map external customer impact on internal API breakages.

## Requirements
1. **Integration Config**: Allow orgs to provide Stripe API keys and Salesforce credentials in the enterprise settings page.
2. **Customer Mapping**: When a breaking change is detected in an externally-facing API, query Stripe/Salesforce to identify which customers consume that API.
3. **PR Comment Section**: Add a "💳 Customer Impact" section in the PR comment listing the number of affected paying customers and their MRR.
4. **Webhook Security**: Use API key validation for all external webhook endpoints.
