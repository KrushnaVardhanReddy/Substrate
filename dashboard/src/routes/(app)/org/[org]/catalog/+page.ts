import type { PageLoad } from './$types';
import { env } from '$env/dynamic/public';

export const load: PageLoad = async ({ fetch, params }) => {
	const org = params.org;
	const isBrowser = typeof window !== 'undefined';
	const baseUrl = isBrowser ? '' : (env.PUBLIC_API_URL || 'http://localhost:8090');
	const token = env.PUBLIC_API_TOKEN || '';

	try {
		// Attempt to fetch from /api/v1/nodes if available.
		// If it's 404, we'll fall back to mock data or /api/v1/graph.
		const response = await fetch(`${baseUrl}/api/v1/nodes`, {
			headers: {
				Authorization: `Bearer ${token}`
			}
		});

		if (response.ok) {
			const nodes = await response.json();
			return { org, nodes: nodes || [] };
		}
	} catch (e) {
		console.error(`Error fetching nodes:`, e);
	}

	// Mock data fallback as requested by the spec for zero-config demo
	const mockNodes = [
		{
			id: "demo-org/core-service",
			name: "core-service",
			owner: "platform",
			protocol: "GraphQL",
			status: "SAFE",
			upstreamCount: 1,
			downstreamCount: 2
		},
		{
			id: "demo-org/gateway-service",
			name: "gateway-service",
			owner: "infrastructure",
			protocol: "OpenAPI",
			status: "SAFE",
			upstreamCount: 1,
			downstreamCount: 2
		},
		{
			id: "demo-org/mobile-ios",
			name: "mobile-ios",
			owner: "mobile",
			protocol: "REST",
			status: "BREAKING",
			upstreamCount: 1,
			downstreamCount: 0
		},
		{
			id: "demo-org/postgres-driver",
			name: "postgres-driver",
			owner: "data",
			protocol: "TCP",
			status: "SAFE",
			upstreamCount: 0,
			downstreamCount: 1
		}
	];

	return { org, nodes: mockNodes };
};
