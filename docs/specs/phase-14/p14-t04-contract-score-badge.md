# P14-T04: Contract Score Badge (Viral Growth Mechanism)

## Overview
A `shields.io`-style embeddable README badge showing a repo's API contract reliability score (e.g., `CONTRACT: A+ | 98% | 0 breaks in 90 days`). This acts as a viral growth mechanism for open-source and internal usage.

## Requirements
1. **Score Calculation Engine**: Implement a function that calculates a letter grade based on breaking change frequency, blast radius, and spec-first adoption rate over the last 90 days.
2. **Badge SVG Generator**: Add an endpoint `GET /api/badges/:org/:repo` that dynamically generates and returns a standard SVG image with the calculated score.
3. **Caching**: Cache the SVG output aggressively (e.g., 24-hour TTL) to prevent DB hammering from GitHub README views.
4. **Dashboard Snippet**: In the repo view, provide a copyable markdown snippet `[![Contract Score](...)]` for developers to paste into their READMEs.
