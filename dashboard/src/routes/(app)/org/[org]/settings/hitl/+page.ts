// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

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
