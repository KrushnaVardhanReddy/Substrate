# P15-T15: Air-Gapped License Validator

## Overview
Implement an air-gapped cryptographically signed license validation system for On-Premise VPC Enterprise deployments. Allows the Substrate engine to read a `.lic` (JWT) file, verify the signature, and enforce features/expiry.

## Requirements
1. **License Format**: Use JWTs signed with RS256. The payload should contain `org_name`, `expires_at`, and `features` (an array of strings).
2. **Validator Middleware**: Add a Go HTTP middleware that intercepts requests for premium features and checks the active license state.
3. **Graceful Degradation**: If the license is missing or expired, do not panic or crash. Instead, fall back to "Free Tier" behavior.
4. **Viper Integration**: The engine should read the license key path from the `SUBSTRATE_LICENSE_FILE` environment variable via Viper.
