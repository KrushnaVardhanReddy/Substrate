// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import type { PageLoad } from './$types';
import { env } from '$env/dynamic/public';

export const load: PageLoad = async ({ fetch, params }) => {
	const org = params.org;
	const apiUrl = env.PUBLIC_API_URL || 'http://localhost:8090';
	const token = env.PUBLIC_API_TOKEN || 'local-dev-token';

	// Procedurally generate 200 nodes and 600 edges for the stress-test org demo
	if (org === 'stress-test') {
		const json = [];
		for (let i = 0; i < 199; i++) {
			json.push({ provider: `node-${i}`, consumer: `node-${i + 1}`, status: 'SAFE' });
		}
		for (let i = 0; i < 600; i++) {
			const providerIdx = Math.floor(Math.random() * 198);
			const consumerIdx = providerIdx + 1 + Math.floor(Math.random() * (199 - providerIdx));
			json.push({ provider: `node-${providerIdx}`, consumer: `node-${consumerIdx}`, status: Math.random() > 0.95 ? 'BREAKING' : 'SAFE' });
		}
		return { org, graphData: json };
	}

	// Use relative URL in browser so Playwright can intercept it easily
	const baseUrl = env.PUBLIC_API_URL || 'http://localhost:8090';
	
	try {
		const response = await fetch(`${baseUrl}/api/v1/graph/${org}`, {
			headers: {
				Authorization: `Bearer ${token}`
			}
		});

		if (response.ok) {
			const graphData = await response.json();
			return { org, graphData: graphData || [] };
		} else {
			console.error(`Failed to fetch graph: ${response.statusText}`);
		}
	} catch (e) {
		console.error(`Error fetching graph:`, e);
	}

	// Fallback mock data removed
	return {
		org,
		graphData: []
	};
};
