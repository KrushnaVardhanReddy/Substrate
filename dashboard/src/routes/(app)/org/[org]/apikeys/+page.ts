import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params, fetch }) => {
	const res = await fetch(`/api/v1/org/${params.org}/apikeys`);

	if (!res.ok) {
		return {
			keys: []
		};
	}

	const keys = await res.json();
	return {
		keys: keys || []
	};
};
