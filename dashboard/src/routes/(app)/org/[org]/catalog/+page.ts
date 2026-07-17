import type { PageLoad } from './$types';
import { env } from '$env/dynamic/public';

export const load: PageLoad = async ({ fetch, params }) => {
	const org = params.org;

	// Fallback mock data for demo purposes if backend fails
	let catalogData = [
		{
			id: "demo-org/core-service",
			name: "core-service",
			owner: "platform",
			protocol: "gRPC",
			status: "SAFE",
		},
		{
			id: "demo-org/postgres-driver",
			name: "postgres-driver",
			owner: "data",
			protocol: "TCP",
			status: "SAFE",
		},
		{
			id: "demo-org/gateway-service",
			name: "gateway-service",
			owner: "infrastructure",
			protocol: "OpenAPI",
			status: "SAFE",
		},
		{
			id: "demo-org/mobile-ios",
			name: "mobile-ios",
			owner: "mobile",
			protocol: "GraphQL",
			status: "BREAKING",
		}
	];

	const isBrowser = typeof window !== 'undefined';
	const baseUrl = isBrowser ? '' : (env.PUBLIC_API_URL || 'http://localhost:8090');
	const token = env.PUBLIC_API_TOKEN || '';

	try {
		const response = await fetch(`${baseUrl}/api/v1/nodes?org=${org}`, {
			headers: {
				Authorization: `Bearer ${token}`
			}
		});

		if (response.ok) {
			const apiNodes = await response.json();
			if (apiNodes && apiNodes.length > 0) {
				catalogData = apiNodes.map((node: any) => {
					// We construct fields based on typical schema if available
					return {
						id: node.id || node.name,
						name: (node.id || node.name).split('/').pop(),
						owner: node.metadata?.team || 'unknown',
						protocol: node.metadata?.type === 'openapi' ? 'OpenAPI' : (node.metadata?.type || 'unknown'),
						status: node.status || 'SAFE'
					}
				});
			}
		} else {
			console.error(`Failed to fetch nodes: ${response.statusText}`);
		}
	} catch (e) {
		console.error(`Error fetching nodes:`, e);
	}

	return { org, catalogData };
};
