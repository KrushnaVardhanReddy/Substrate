// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params, fetch }) => {
	// Future phase will fetch schema content from API.
	// For now, return dummy data for UI display.
	let canDeploy = true;
	let canRollback = true;

	try {
		const res = await fetch(`/api/v1/impact/${params.org}/${params.repo}`);
		if (res.ok) {
			const impactData = await res.json();
			canDeploy = impactData.can_deploy ?? true;
			canRollback = impactData.can_rollback ?? true;
		}
	} catch (e) {
		console.warn("Failed to fetch impact data", e);
	}

	return {
		repoName: params.repo,
		orgName: params.org,
		canDeploy,
		canRollback,
		schemaContent: `openapi: 3.0.0
info:
  title: Dummy API
  version: 1.0.0
paths:
  /users:
    get:
      summary: Returns a list of users.
      responses:
        '200':
          description: A JSON array of user names
          content:
            application/json:
              schema:
                type: array
                items:
                  type: string
`
	};
};
