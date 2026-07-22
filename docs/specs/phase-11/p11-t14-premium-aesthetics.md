# Spec: P11-T14 - Premium Aesthetics System

## 1. Overview
Overhaul the Substrate UI to align with top-tier Enterprise SaaS aesthetics. This includes a deep dark mode, glassmorphism overlays, curated HSL color palettes, and edge-flow micro-animations.

## 2. Requirements
- Define a global CSS variable system in `src/app.css` (e.g., `--bg-deep: #0F1117`).
- Implement Glassmorphism utility classes (`bg-opacity-20 backdrop-blur-md`).
- Update global typography to a modern sans-serif (Inter or outfit).
- Update primary buttons and interactive elements with subtle hover/focus micro-animations.

## 3. Stitch & Jules Workflow
- **Stitch:** Generate a static HTML/CSS mockup of the new base layout and typography.
- **Jules:** Extract the CSS variables and structural changes from the mockup and apply them to the SvelteKit global layout (`src/routes/+layout.svelte` and `app.css`).
