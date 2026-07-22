# P15-T11: Substrate for Startups (Free Tier with Public Audit Trail)

## Overview
Free forever for OSS projects with a public API reliability profile at `substrate.io/profile/myorg/api`. Startups use the profile in enterprise sales as proof of API stability.

## Requirements
1. **Public Profile Page**: Add a publicly accessible SvelteKit route `/(public)/profile/[org]` showing the org's API health history.
2. **OSS Tier Flag**: Add an `is_oss` boolean flag to the `organizations` table gating public profile features.
3. **Reliability Metrics**: Show a 90-day rolling chart of: breaking changes per week, blast radius trend, and uptime percentage.
4. **Shareable URL**: Provide a one-click copy button for the public profile URL.
