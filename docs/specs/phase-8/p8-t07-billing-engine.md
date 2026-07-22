# Spec: P8-T07 - Billing & Subscription Engine (Paywall Pause)

## 1. Overview
Before launching on the GitHub Marketplace, Substrate must have a mechanism to enforce the commercial model. We will implement a 90-day trial period, after which the analysis is paused until a valid Stripe subscription is linked to the GitHub organization.

## 2. Requirements

### 2.1 Database Tracking
- Add `trial_ends_at TIMESTAMPTZ` and `stripe_customer_id TEXT` to the `organizations` table.
- When an organization is first created via the GitHub App installation, `trial_ends_at` is set to `NOW() + INTERVAL '90 days'`.

### 2.2 Trial Enforcement
- Update the GitHub App webhook receiver (`github-app/src/index.ts` or Go API equivalent):
  - Check if `NOW() > trial_ends_at` and `stripe_customer_id IS NULL`.
  - If true, **DO NOT run analysis**.
  - Instead, immediately create a PR check run with a `neutral` or `action_required` status, stating:
    `Substrate 90-day trial has expired. Please visit the dashboard to upgrade to the Enterprise plan and resume API protection.`
- This acts as a "Graceful Pause". It does not block PRs by returning a failure (exit 1/2), to prevent locking up customer CI/CD pipelines unexpectedly.

### 2.3 Stripe Webhooks
- Expose a `/api/v1/webhooks/stripe` endpoint.
- Listen for `checkout.session.completed` and `customer.subscription.deleted`.
- Update the `organizations` table with `stripe_customer_id` when they subscribe.

## 3. Implementation Steps
1. Add migration `0009_billing_fields.up.sql` to modify the `organizations` table.
2. Integrate `stripe-go` to verify incoming webhook signatures and update DB status.
3. Modify the PR webhook handler to check billing status before executing diff analysis.
4. Provide a clear, actionable error message in the GitHub Check Run linking to the Stripe checkout page.
