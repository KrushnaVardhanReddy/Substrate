import { describe, it, expect, vi } from 'vitest';
import { processDeprecations } from '../src/deprecation';
import * as githubClient from '../src/github-client';
import { CrossRepoCheckResponse, Deprecation } from '../src/types';

describe('Deprecation Campaigns', () => {
  it('should open issues for consumers that use the endpoint, and close if usage is 0', async () => {
    const createIssueMock = vi.spyOn(githubClient, 'createIssue').mockResolvedValue({});
    const findOpenIssueMock = vi.spyOn(githubClient, 'findOpenIssue').mockImplementation(async (token, owner, repo) => {
      if (repo === 'repo1') {
        return { items: [] }; // No open issue
      }
      if (repo === 'repo2') {
        return { items: [{ number: 123 }] }; // Has open issue
      }
      return { items: [] };
    });
    const closeIssueMock = vi.spyOn(githubClient, 'closeIssue').mockResolvedValue();

    const deprecations: Deprecation[] = [
      { endpoint: '/api/v1/users', sunset_date: '2024-12-31' },
    ];

    const crossRepoResponse: CrossRepoCheckResponse = {
      total_consumers: 2,
      broken_consumers: 0,
      is_safe: true,
      results: [
        {
          consumer_repo: 'org1/repo1',
          status: 'warning',
          diff_report: {
            breaking_changes: [],
            warnings: [
              { rule_id: 'ENDPOINT_DEPRECATED', path: '/api/v1/users', description: '' }
            ],
            safe_changes: [],
            summary: { breaking_count: 0, warning_count: 1, info_count: 0 }
          }
        },
        {
          consumer_repo: 'org2/repo2',
          status: 'safe',
          diff_report: {
            breaking_changes: [],
            warnings: [], // No usage
            safe_changes: [],
            summary: { breaking_count: 0, warning_count: 0, info_count: 0 }
          }
        }
      ]
    };

    await processDeprecations('fake-token', deprecations, crossRepoResponse);

    expect(createIssueMock).toHaveBeenCalledTimes(1);
    expect(createIssueMock).toHaveBeenCalledWith(
      'fake-token',
      'org1',
      'repo1',
      'Deprecation Notice: /api/v1/users',
      expect.stringContaining('2024-12-31')
    );

    expect(findOpenIssueMock).toHaveBeenCalledTimes(2);

    expect(closeIssueMock).toHaveBeenCalledTimes(1);
    expect(closeIssueMock).toHaveBeenCalledWith(
      'fake-token',
      'org2',
      'repo2',
      123,
      expect.stringContaining('dropped to 0%')
    );

    createIssueMock.mockRestore();
    findOpenIssueMock.mockRestore();
    closeIssueMock.mockRestore();
  });
});
