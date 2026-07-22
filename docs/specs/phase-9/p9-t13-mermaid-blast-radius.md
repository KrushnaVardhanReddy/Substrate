# P9-T13: Embedded Mermaid Blast Radius

## Overview
Upgrade the GitHub PR comment bot to render a visual Mermaid.js flowchart of the exact blast radius directly inside the PR.

## Requirements
1. **Mermaid Generation**: After blast radius traversal, generate a Mermaid `flowchart TD` string showing provider → consumer edges.
2. **GitHub Embedding**: Wrap the Mermaid code in a fenced ` ```mermaid ``` ` block inside the PR comment (GitHub renders this natively).
3. **Size Limit**: Cap graph rendering at 20 nodes to prevent enormous diagrams; show a truncation notice if exceeded.
