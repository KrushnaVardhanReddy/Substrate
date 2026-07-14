# Spec: P11-T11 - Global Command Palette (Cmd+K)

## 1. Overview
Implement a Raycast-style command palette overlay that allows users to instantly search for and navigate to specific microservices, teams, or settings across the enterprise graph.

## 2. Requirements
- Listen for the `Cmd+K` (or `Ctrl+K`) keyboard shortcut globally.
- Render a blurred, centered modal overlay with a text input.
- Fuzzy search across the loaded graph data (nodes).
- Keyboard navigation (Up/Down arrows, Enter to select).
- Selecting a node triggers an event to focus that node in the main graph view.

## 3. Stitch & Jules Workflow
- **Stitch:** Generate a mockup of the Command Palette modal overlay (dark theme, glassmorphism, search results list).
- **Jules:** Create the `CommandPalette.svelte` component, wire up keyboard event listeners, and implement fuzzy search logic.
