import type { PageLoad } from './$types';
import { env } from '$env/dynamic/public';

export const load: PageLoad = async ({ fetch, params }) => {
	const org = params.org;
	const apiUrl = env.PUBLIC_API_URL || 'http://localhost:8090';
	const token = env.PUBLIC_API_TOKEN || '';

	try {
		const response = await fetch(`${apiUrl}/api/v1/repos/${org}`, {
			headers: {
				Authorization: `Bearer ${token}`
			}
		});

		if (response.ok) {
			const repos = await response.json();
			return { repos: repos || [] };
		} else {
			console.error(`Failed to fetch repos: ${response.statusText}`);
			return { repos: [] };
		}
	} catch (e) {
		console.error(`Error fetching repos:`, e);
		return { repos: [] };
	}
};
