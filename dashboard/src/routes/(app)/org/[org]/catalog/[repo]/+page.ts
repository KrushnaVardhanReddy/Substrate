// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import type { PageLoad } from './$types';

import { env } from '$env/dynamic/public';

export const load: PageLoad = async ({ fetch, params }) => {
	const org = params.org;
	const repo = params.repo;
	const isBrowser = typeof window !== 'undefined';
	const baseUrl = isBrowser ? '' : (env.PUBLIC_API_URL || 'http://localhost:8090');
	const token = env.PUBLIC_API_TOKEN || '';

	let repoData = null;
	try {
		const response = await fetch(`${baseUrl}/api/v1/repos/${org}`, {
			headers: { Authorization: `Bearer ${token}` }
		});
		if (response.ok) {
			const repos = await response.json();
			repoData = repos.find((r: any) => r.name === repo || r.full_name === `${org}/${repo}`) || null;
		}
	} catch (e) {
		console.error('Error fetching repos:', e);
	}

	// According to spec, Substrate stores raw schema strings in DB.
	// For OpenAPI specs, we render with @stoplight/elements using raw YAML.
	// Provide a dummy OpenAPI YAML string for testing this functionality.
	const dummyYaml = `openapi: 3.0.0
info:
  title: ${repo} API
  version: 1.0.0
  description: Automatically generated API documentation for ${repo}.
paths:
  /users:
    get:
      summary: Get all users
      responses:
        '200':
          description: A list of users
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object
                  properties:
                    id:
                      type: string
                    name:
                      type: string
`;

	const apiBaseUrl = import.meta.env.VITE_API_URL || 'https://api.substrate.com';
	const dashboardUrl = import.meta.env.VITE_DASHBOARD_URL || 'https://app.substrate.com';

	return { org, repo, yamlString: dummyYaml, apiBaseUrl, dashboardUrl, repoData };
};
