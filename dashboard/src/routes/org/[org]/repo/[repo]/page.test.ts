import { describe, it, expect } from 'vitest';
import { load } from './+page';

describe('Repo Page Load Function', () => {
	it('returns dummy schema data successfully', async () => {
		const result = await load({ params: { org: 'testorg', repo: 'testrepo' } } as any);

		expect((result as any).orgName).toBe('testorg');
		expect((result as any).repoName).toBe('testrepo');
		expect((result as any).schemaContent).toContain('openapi: 3.0.0');
	});
});
