import { browser } from '$app/environment';
import { env } from '$env/dynamic/public';

/**
 * Tracks a named event with optional properties.
 * This is a no-op unless PUBLIC_POSTHOG_KEY is set.
 * Telemetry MUST NEVER crash the application.
 */
export async function trackEvent(name: string, props?: Record<string, unknown>): Promise<void> {
  if (!browser) return;

  // We also want to support sending to our internal backend /api/v1/telemetry/track endpoint
  try {
    const isEnabled = localStorage.getItem('telemetryEnabled') !== 'false';
    if (isEnabled) {
        await fetch('/api/v1/telemetry/track', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ event: name, properties: props })
        });
    }
  } catch {
      // Silently swallow network failures to prevent UI crashing
  }

  if (!env.PUBLIC_POSTHOG_KEY) return;
  try {
    // posthog is loaded via CDN script tag in app.html
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (window as any).posthog?.capture(name, props ?? {});
  } catch {
    // Silently swallow — telemetry must never crash the app
  }
}
