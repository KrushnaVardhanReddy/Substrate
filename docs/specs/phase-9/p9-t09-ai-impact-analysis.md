# P9-T09: AI Impact Analysis Summaries

## Overview
Pass cross-repo blast radius checks to the AI handler to generate a plain-English impact summary on PRs.

## Requirements
1. **Data Input**: After blast radius traversal, collect the list of affected downstream repos and the nature of each breaking change.
2. **AI Prompt**: Construct a structured prompt to the LLM summarizing the breaking changes and asking for a concise risk narrative.
3. **PR Comment Section**: Append an "🤖 AI Impact Analysis" section at the bottom of the GitHub PR comment with the generated narrative.
4. **Max Tokens**: Limit the AI response to 200 tokens to keep the PR comment concise.
