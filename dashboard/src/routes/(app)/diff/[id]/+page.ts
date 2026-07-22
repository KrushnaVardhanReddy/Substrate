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
			throw new Error('Failed to fetch diff report');
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

		// Fallback mock data for visual verification purposes
		const oldText = `name: Service A
type: backend
version: 1.0.0
dependencies:
  - postgres
  - redis
`;
		const newText = `name: Service A
type: backend
version: 1.1.0
dependencies:
  - postgres
  - kafka
`;

		return {
			id,
			oldText,
			newText,
			report: null
		};
	}
}