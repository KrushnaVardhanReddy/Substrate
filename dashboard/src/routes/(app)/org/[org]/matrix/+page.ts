import type { PageLoad } from './$types';
import { env } from '$env/dynamic/public';

export const load: PageLoad = async ({ fetch, params }) => {
	const org = params.org;
	const apiUrl = env.PUBLIC_API_URL || 'http://localhost:8090';
	const token = env.PUBLIC_API_TOKEN || 'local-dev-token';

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

	// Fallback mock data removed
	return {
		org,
		matrixData: { providers: [], consumers: [], grid: [] }
	};
};
