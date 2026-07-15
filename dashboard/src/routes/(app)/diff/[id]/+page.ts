import { error } from '@sveltejs/kit';
import { env } from '$env/dynamic/public';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params, fetch }) => {
	const apiUrl = env.PUBLIC_API_URL || 'http://localhost:8090';
	const id = params.id;

	try {
		const res = await fetch(`${apiUrl}/api/v1/diff/${id}`);
		if (!res.ok) {
			if (res.status === 404) {
				// We don't want to throw an error for mock rendering, we will handle it in frontend
				return { id, oldText: '', newText: '', report: null };
			}
			return { id, oldText: '', newText: '', report: null };
		}

		const report = await res.json();

		let oldText = '';
		let newText = '';

        if (report.base_schema && report.head_schema) {
            oldText = report.base_schema;
            newText = report.head_schema;
        } else {
            oldText = JSON.stringify(report, null, 2);
            newText = JSON.stringify(report, null, 2);
        }

		return {
			id,
			oldText,
			newText,
            report
		};
	} catch (e) {
        console.error('Error fetching diff:', e);
        return { id, oldText: '', newText: '', report: null };
	}
};
