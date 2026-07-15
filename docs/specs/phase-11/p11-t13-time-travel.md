# Spec: P11-T13 - Time-Travel Graph Replay

## 1. Overview
Add a timeline scrubber (slider) to the bottom of the Svelte Flow graph visualization. This allows enterprise architects to drag a slider back in time and view the exact state of their microservice dependencies on any historical date.

## 2. Requirements
- Create `TimeTravelScrubber.svelte` as a floating component anchored to the bottom-center of the graph view.
- It must contain a Play/Pause button, a slider input (`type="range"`), and text indicating the selected date.
- Hook the slider up to a Svelte 5 `$state` variable representing the `currentTimestamp`.
- Ensure it uses the Premium Aesthetics system (Glassmorphism, Dark Mode `var(--bg-card)` with `backdrop-filter`, and `var(--accent)`).
- Integrate the scrubber into `dashboard/src/routes/(app)/org/[org]/graph/+page.svelte`.

## 3. Stitch & Jules Workflow
- **Stitch:** Generate a high-fidelity mockup of the scrubber bar floating over the dark canvas.
- **Jules:** Implement the `TimeTravelScrubber.svelte` and integrate it into the graph page using the mockup as the source of truth for UI/UX.
