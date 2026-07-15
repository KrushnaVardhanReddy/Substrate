import { error } from '@sveltejs/kit';
import { env } from '$env/dynamic/public';

export async function load({ params, fetch }) {
	const apiUrl = env.PUBLIC_API_URL || 'http://localhost:8090';
	const id = params.id;

	try {
		const res = await fetch(`${apiUrl}/api/v1/diff/${id}`);
		if (!res.ok) {
			if (res.status === 404) {
				error(404, 'Diff report not found');
			}
			error(500, 'Failed to fetch diff report');
		}

		const report = await res.json();
		// Extract oldText and newText from report if available. The spec doesn't detail what DiffReport will look like, so we format it for display.
		// If base/head are not returned, we can format the breaking changes array as old/new representation.

		// Fallback to empty text if oldText/newText isn't present
		let oldText = '';
		let newText = '';

        if (report.base_schema && report.head_schema) {
            oldText = report.base_schema;
            newText = report.head_schema;
        } else {
            // Alternatively stringify the full report to view it
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
        if (e instanceof Error) {
            console.error('Error fetching diff:', e);
            error(500, 'Failed to connect to backend');
        }
        throw e;
	}
}
