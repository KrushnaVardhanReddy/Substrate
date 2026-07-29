import { env } from '$env/dynamic/public';
import type { PageLoad } from './$types';


export const load: PageLoad = async ({ params, fetch }) => {
	const { org } = params;

	const res = await fetch(`${env.PUBLIC_API_URL}/api/v1/mcp/hitl-queue/${org}`, {

	});

	let items = [];
	if (res.ok) {
		items = await res.json();
	}

	return {
		org,
		items
	};
};
