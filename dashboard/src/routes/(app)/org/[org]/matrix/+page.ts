import type { PageLoad } from './$types';
import { env } from '$env/dynamic/public';

export const load: PageLoad = async ({ fetch, params }) => {
	const org = params.org;
	const apiUrl = env.PUBLIC_API_URL || 'http://localhost:8090';
	const token = env.PUBLIC_API_TOKEN || '';

	try {
		const response = await fetch(`${apiUrl}/api/v1/matrix/${org}`, {
			headers: {
				Authorization: `Bearer ${token}`
			}
		});

		if (response.ok) {
			const matrixData = await response.json();
			return { org, matrixData: matrixData || [] };
		} else {
			console.error(`Failed to fetch matrix: ${response.statusText}`);
		}
	} catch (e) {
		console.error(`Error fetching matrix:`, e);
	}

	// Mock data for the matrix
	return {
		org,
		matrixData: {
			providers: ["users-api", "checkout-api"],
			consumers: ["frontend-web (PROD)", "mobile-ios (V1.2)", "mobile-android (V1.1)"],
			grid: [
				{
					provider: "users-api",
					versions: [
						{
							version: "v2.1.0",
							results: ["COMPATIBLE", "COMPATIBLE", "INCOMPATIBLE"]
						},
						{
							version: "v2.0.0",
							results: ["COMPATIBLE", "COMPATIBLE", "COMPATIBLE"]
						}
					]
				},
				{
					provider: "checkout-api",
					versions: [
						{
							version: "v1.5.2",
							results: ["COMPATIBLE", "UNKNOWN", "COMPATIBLE"]
						}
					]
				}
			]
		}
	};
};
