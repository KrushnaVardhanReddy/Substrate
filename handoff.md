# Substrate — Shift Handoff Document

> **Date:** 2026-07-25
> **Current Focus:** Fixing Dependency Graph Frontend Rendering (Svelte 5 + Cytoscape)

## ✅ Completed This Session

1. **Dependency Graph Rendering Engine Migration**
   - ✅ **Issue:** The dependency graph on the frontend (`/org/[org]/graph`) was completely blank. This was caused by `SvelteFlow` v1.6.2 throwing `lifecycle_outside_component` context tracker errors in Svelte 5, especially when combined with async data loading in SvelteKit's declarative routing.
   - ✅ **Fix:** We ripped out the incompatible `SvelteFlow` library and reverted the rendering engine back to **Cytoscape** and **cytoscape-dagre**, which the project originally used 10 days ago.
   - ✅ **Integration:** Integrated Cytoscape directly with the new Server-Sent Events (SSE) logic natively in Svelte 5 using the `$effect` rune.
   - ✅ **Layout:** Fixed CSS layout overlap bugs by setting the `cyContainer` flex wrapper to `flex: 1` rather than absolute positioning, allowing it to seamlessly fit below the header.
   - ✅ **Styling:** Updated the node padding and sizing styles (`width: 'auto'`, `height: 'auto'`) to conform to modern Cytoscape standards (preventing text bleed and deprecation warnings).
   - ✅ **Clean up:** Removed the manual `getLayoutedElements` DAGRE math function (since Cytoscape computes layout natively), cleaned up all `local-dev-token` hardcoding, and purged all temporary API proxy endpoints/debug files created during the debugging session.

---

## 🔁 Next Steps for the Evening Shift

1. **Verify Backend/Frontend Synchronization:** Ensure the graph properly interacts with the newly verified MCP parity server.
2. **Review Node Rendering:** Check if the nodes correctly reflect the live DB state.
