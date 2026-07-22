# P15-T09: "Substrate Certified" Partner Program

## Overview
API platform vendors (Kong, AWS API Gateway, Apigee, Cloudflare) pay $2k–$20k/year for certified native integration status.

## Requirements
1. **Partner Registry**: Add a `partner_integrations` DB table tracking certified vendor integrations.
2. **Admin Dashboard**: Create an internal admin page (protected route) to manage certified partners.
3. **Public Badge**: Add a "Substrate Certified" badge component displayed on the partner's integration page in the Substrate dashboard.
4. **Integration Verification**: Implement a webhook validation handshake that a partner must complete to achieve "Certified" status.
