// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

<script lang="ts">
	let { grade, org, repo } = $props<{ grade: string, org: string, repo: string }>();

	function getGradeClass(g: string) {
		switch (g) {
			case 'A': return 'bg-green-100 text-green-800 border border-green-200';
			case 'B': return 'bg-teal-100 text-teal-800 border border-teal-200';
			case 'C': return 'bg-yellow-100 text-yellow-800 border border-yellow-200';
			case 'D': return 'bg-orange-100 text-orange-800 border border-orange-200';
			case 'F': return 'bg-red-100 text-red-800 border border-red-200';
			default: return 'bg-gray-100 text-gray-800 border border-gray-200';
		}
	}
</script>

<a href="/org/{org}/scorecard/{repo}" data-testid="health-grade-badge" class="grade-badge inline-flex items-center justify-center w-8 h-8 rounded font-bold text-sm {getGradeClass(grade)} hover:opacity-80 transition-opacity">
	{grade}
</a>
