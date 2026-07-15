# Spec: P11-T07 - Visual API Design Studio

## 1. Overview
A Drag-and-drop OpenAPI designer built directly into the Substrate UI to empower PMs and Architects to design APIs before coding.

## 2. Requirements
- Create `dashboard/src/routes/(app)/studio/+page.svelte` (a new page in the dashboard).
- The studio should have a split view: left side is a JSON/YAML editor (using a lightweight textarea or simple editor), right side is a visual representation of endpoints (e.g., GET /users).
- When the user types in the editor, parse the YAML (using `js-yaml` which you must install or mock) and render the endpoints.
- Allow adding a new endpoint via a simple UI form, which updates the YAML.
- Must follow the Premium Aesthetics design system (dark mode, glassmorphism).

## 3. Stitch & Jules Workflow
- **Jules:** Implement the Studio page and the bidirectional data binding between the YAML text and the visual endpoint cards.
