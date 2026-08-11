// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params, fetch }) => {
	const org = params.org;
	const repo = params.repo;

	let specString = null;
	try {
		const res = await fetch(`/api/v1/spec/${org}/${repo}`);
		if (res.ok) {
			const data = await res.json();
			specString = data.spec;
		}
	} catch (e) {
		console.error("Failed to fetch API spec:", e);
	}

	return { org, repo, specString };
};
