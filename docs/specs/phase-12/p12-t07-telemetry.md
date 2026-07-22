# Spec: P12-T07 — Telemetry & Crash Reporting

## 1. Overview
Integrate PostHog (privacy-first, self-hostable) into the SvelteKit frontend behind an environment variable guard so production errors and key user interactions are tracked without impacting dev/test environments.

## 2. Owner
**Stitch** (Frontend)

## 3. Files Modified
- `dashboard/src/lib/utils/telemetry.ts` ← **new file**
- `dashboard/src/app.html` ← add PostHog script snippet
- `dashboard/src/routes/(app)/org/[org]/graph/+page.svelte` ← add 1 `trackEvent()` call

## 4. Requirements

### `telemetry.ts` — Wrapper Utility
```typescript
// dashboard/src/lib/utils/telemetry.ts
import { browser } from '$app/environment';
import { env } from '$env/dynamic/public';

export function trackEvent(name: string, props?: Record<string, unknown>): void {
  if (!browser) return;
  if (!env.PUBLIC_POSTHOG_KEY) return; // No-op in dev and test
  try {
    // @ts-expect-error posthog is loaded via script tag in app.html
    window.posthog?.capture(name, props);
  } catch {
    // Silently swallow — telemetry must never crash the app
  }
}
```

### `app.html` — PostHog Bootstrap
Add the PostHog JS snippet **only** when `%sveltekit.head%` is present and guard with a Vite env check. Use `posthog-js` CDN snippet pattern:
```html
<!-- Only loads if PUBLIC_POSTHOG_KEY is set at build time -->
%sveltekit.head%
```
Insert the PostHog init script as a `<script>` tag that reads `window.__POSTHOG_KEY__` (injected via a SvelteKit hook or just leave it as a manual env-gated snippet).

> **Simplest acceptable approach:** Add the PostHog snippet commented with `<!-- PostHog Telemetry — only active in production -->` and wrap the init call in `if (window.__POSTHOG_KEY__)`.

### `+page.svelte` — Event Tracking
In the `onnodeclick` handler, add:
```typescript
import { trackEvent } from '$lib/utils/telemetry';
// ...inside onnodeclick:
trackEvent('node_clicked', { nodeId: node.id, nodeType: node.data.type });
```

## 5. Technical Constraints
- `trackEvent` MUST be a no-op when `PUBLIC_POSTHOG_KEY` is not set (protects all existing tests).
- The PostHog script tag in `app.html` must NOT break the existing `dashboard.spec.ts` E2E test.
- Do not import `posthog-js` as an npm package — use the CDN snippet approach to keep bundle size zero.
- The `+page.svelte` change is a single 1-line addition in the existing `onnodeclick` callback — do not refactor the surrounding code.

## 6. Success Criteria
- `npx playwright test` continues to pass all 11 existing tests unchanged.
- `dashboard/src/lib/utils/telemetry.ts` exists and exports `trackEvent`.
- In production builds with `PUBLIC_POSTHOG_KEY` set, PostHog initializes and `node_clicked` events are captured.
