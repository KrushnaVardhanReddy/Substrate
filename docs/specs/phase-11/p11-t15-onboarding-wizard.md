# Spec: P11-T15 - Zero-to-One Onboarding Wizard

## 1. Overview
Create a frictionless, animated onboarding flow for first-time enterprise users: Connect GitHub -> Scan Repositories -> Build Graph.

## 2. Requirements
- Create a new route: `src/routes/onboarding/+page.svelte`.
- Step 1: "Connect GitHub" (Displays GitHub App installation link).
- Step 2: "Scanning Repositories" (Progress bar simulating the background River queue tasks).
- Step 3: "Building Graph" (Animated graph construction preview).
- Final Step: "Enter Dashboard" button that redirects to `/`.
- Smooth cross-fade transitions between steps.

## 3. Stitch & Jules Workflow
- **Stitch:** Generate mockups for the 3 visual states of the onboarding wizard.
- **Jules:** Implement the multi-step Svelte component flow, utilizing Svelte's built-in transition APIs (`fade`, `slide`).
