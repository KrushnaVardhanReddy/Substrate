import type { PageLoad } from './$types';
import { env } from '$env/dynamic/public';

export const load: PageLoad = async ({ fetch, params }) => {
	const org = params.org;
	const apiUrl = env.PUBLIC_API_URL || 'http://localhost:8090';
	const token = env.PUBLIC_API_TOKEN || '';

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
	const isBrowser = typeof window !== 'undefined';
	const baseUrl = isBrowser ? '' : (env.PUBLIC_API_URL || 'http://localhost:8090');
	
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

	// Fallback mock data for demo purposes if backend fails
	return {
		org,
		graphData: [
			{ provider: "core/auth", consumer: "api/gateway-service", status: "SAFE" },
			{ provider: "db/postgres-driver", consumer: "api/gateway-service", status: "SAFE" },
			{ provider: "utils/logger", consumer: "api/gateway-service", status: "BREAKING" },
			{ provider: "api/gateway-service", consumer: "frontend/dashboard", status: "SAFE" },
			{ provider: "api/gateway-service", consumer: "workers/indexer", status: "SAFE" }
		]
	};
};
