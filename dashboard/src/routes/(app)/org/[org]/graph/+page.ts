import type { PageLoad } from './$types';
import { env } from '$env/dynamic/public';

export const load: PageLoad = async ({ fetch, params }) => {
	const org = params.org;
	const apiUrl = env.PUBLIC_API_URL || 'http://localhost:8090';
	const token = env.PUBLIC_API_TOKEN || '';

	try {
		const response = await fetch(`${apiUrl}/api/v1/graph/${org}`, {
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
