import { describe, it, expect, vi, beforeEach } from 'vitest';
import { processAutoDiscovery } from '../src/discovery.js';
import * as githubClient from '../src/github-client.js';

vi.mock('../src/github-client.js', () => ({
  generateInstallationToken: vi.fn(),
  checkFileExists: vi.fn(),
  getDefaultBranch: vi.fn(),
  getBranchSha: vi.fn(),
  createBranch: vi.fn(),
  createCommitWithFile: vi.fn(),
  createPullRequest: vi.fn()
}));

describe('processAutoDiscovery', () => {
  const env: any = {
    GITHUB_APP_ID: '123',
    GITHUB_APP_PRIVATE_KEY: 'key'
  };

  const repos = [
    {
      id: 1,
      name: 'test-repo',
      full_name: 'test-owner/test-repo',
      owner: { login: 'test-owner', id: 2 }
    }
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(githubClient.generateInstallationToken).mockResolvedValue('fake-token');
    vi.mocked(githubClient.getDefaultBranch).mockResolvedValue('main');
    vi.mocked(githubClient.getBranchSha).mockResolvedValue('fake-sha');
    vi.mocked(githubClient.createPullRequest).mockResolvedValue(42);
  });

  it('should skip if no spec files are found', async () => {
    vi.mocked(githubClient.checkFileExists).mockResolvedValue(false);

    await processAutoDiscovery(env, 123, repos);

    expect(githubClient.checkFileExists).toHaveBeenCalledWith('fake-token', 'test-owner', 'test-repo', 'openapi.yaml', 'main');
    expect(githubClient.createBranch).not.toHaveBeenCalled();
  });

  it('should skip if substrate.yaml already exists', async () => {
    vi.mocked(githubClient.checkFileExists).mockImplementation(async (token, owner, repo, path) => {
      if (path === 'openapi.yaml') return true;
      if (path === 'substrate.yaml') return true;
      return false;
    });

    await processAutoDiscovery(env, 123, repos);

    expect(githubClient.createBranch).not.toHaveBeenCalled();
  });

  it('should create branch, commit, and open PR if spec exists', async () => {
    vi.mocked(githubClient.checkFileExists).mockImplementation(async (token, owner, repo, path) => {
      if (path === 'schema.graphql') return true;
      if (path === 'substrate.yaml') return false;
      return false;
    });

    await processAutoDiscovery(env, 123, repos);

    expect(githubClient.createBranch).toHaveBeenCalledWith('fake-token', 'test-owner', 'test-repo', 'substrate-init', 'fake-sha');
    expect(githubClient.createCommitWithFile).toHaveBeenCalled();
    const commitCall = vi.mocked(githubClient.createCommitWithFile).mock.calls[0];
    expect(commitCall[4]).toBe('substrate.yaml');
    expect(commitCall[5]).toContain('schema_type: graphql');
    expect(commitCall[5]).toContain('base_schema: schema.graphql');
    expect(githubClient.createPullRequest).toHaveBeenCalled();
  });
});
