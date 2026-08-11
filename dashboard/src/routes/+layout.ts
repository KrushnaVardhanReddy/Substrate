// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { env } from '$env/dynamic/public';
import type { LayoutLoad } from './$types';

export const ssr = false;

export const load: LayoutLoad = async ({ fetch }) => {
	const apiUrl = env.PUBLIC_API_URL;
	const orgName = env.PUBLIC_ORG_NAME;
	const token = env.PUBLIC_API_TOKEN;

	try {
		const response = await fetch(`${apiUrl}/api/v1/repos/${orgName}`, {
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
