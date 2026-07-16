import { browser } from '$app/environment';
import { env } from '$env/dynamic/public';

/**
 * Tracks a named event with optional properties.
 * This is a no-op unless PUBLIC_POSTHOG_KEY is set.
 * Telemetry MUST NEVER crash the application.
 */
export function trackEvent(name: string, props?: Record<string, unknown>): void {
  if (!browser) return;
  if (!env.PUBLIC_POSTHOG_KEY) return;
  try {
    // posthog is loaded via CDN script tag in app.html
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (window as any).posthog?.capture(name, props ?? {});
  } catch {
    // Silently swallow — telemetry must never crash the app
  }
}
