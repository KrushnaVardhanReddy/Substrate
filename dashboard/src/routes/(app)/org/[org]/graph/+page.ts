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
			{ 
				provider: "demo-org/core-service", 
				consumer: "demo-org/gateway-service", 
				status: "SAFE",
				provider_metadata: { type: "backend", team: "Platform", databases: ["postgres", "redis"] },
				consumer_metadata: { type: "service", team: "Infrastructure" }
			},
			{ 
				provider: "demo-org/postgres-driver", 
				consumer: "demo-org/core-service", 
				status: "SAFE",
				provider_metadata: { type: "database", team: "DataEng", databases: ["postgres"] },
				consumer_metadata: { type: "backend", team: "Platform", databases: ["postgres", "redis"] }
			},
			{ 
				provider: "demo-org/gateway-service", 
				consumer: "demo-org/frontend-dashboard", 
				status: "SAFE",
				provider_metadata: { type: "service", team: "Infrastructure" },
				consumer_metadata: { type: "frontend", team: "Product" }
			},
			{ 
				provider: "demo-org/gateway-service", 
				consumer: "demo-org/mobile-ios", 
				status: "BREAKING",
				provider_metadata: { type: "service", team: "Infrastructure" },
				consumer_metadata: { type: "mobile", team: "Mobile" }
			}
		]
	};
};
