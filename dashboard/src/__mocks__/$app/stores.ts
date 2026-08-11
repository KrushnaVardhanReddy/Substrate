// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { readable } from 'svelte/store'; export const page = readable({ params: { org: 'myorg' }, url: { pathname: '/org/myorg' } });
