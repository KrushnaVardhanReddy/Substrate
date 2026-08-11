// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

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
