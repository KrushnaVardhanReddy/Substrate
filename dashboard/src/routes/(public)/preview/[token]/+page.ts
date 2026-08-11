// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { error } from '@sveltejs/kit';
import { env } from '$env/dynamic/public';

export async function load({ params, fetch }) {
	const apiUrl = env.PUBLIC_API_URL || 'http://localhost:8090';
	const token = params.token;

	try {
		const res = await fetch(`${apiUrl}/api/v1/preview/${token}`);

		if (!res.ok) {
			if (res.status === 410) {
				return { error: 'This preview link has expired.' };
			}
			if (res.status === 404) {
				return { error: 'Preview not found or invalid token.' };
			}
			throw new Error('Failed to fetch preview report');
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
			token,
			oldText,
			newText,
			report,
			error: null
		};
	} catch (e) {
		console.error('Error fetching preview diff:', e);
		return { error: 'An error occurred while loading the preview.' };
	}
}
