# Spec: P11-T02 - Team Neighborhoods

## 1. Overview
Group microservices visually by team ownership using Cytoscape's "Parent Node" feature, creating bounded sub-flows for teams (e.g., Platform Team, Payments Team).

## 2. Requirements
- Modify the graph data generation to create "Group Nodes" for each unique team.
- Assign the standard microservice nodes as children of these group nodes (`parentNode: 'team-payments'`).
- The Group Node should be a custom Cytoscape node that renders a semi-transparent bounding box with a team label at the top.
- Ensure Dagre layout handles grouped nodes correctly (Dagre supports clustering/compound graphs).

## 3. Stitch & Jules Workflow
- **Jules:** Implement the `GroupNode.svelte` and adapt the Dagre layout to support compound nodes.
