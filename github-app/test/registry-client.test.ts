import { describe, it, expect, vi, beforeEach } from 'vitest';
import { parseConsumersFromYaml, syncToRegistry, crossRepoCheck } from '../src/registry-client.js';

describe('parseConsumersFromYaml', () => {
  it('returns empty array for empty yaml', async () => {
    const res = await parseConsumersFromYaml('');
    expect(res).toEqual([]);
  });

  it('returns empty array for yaml with no consumers key', async () => {
    const res = await parseConsumersFromYaml('base_schema: api/openapi.yaml\nhead_schema: api/openapi.yaml');
    expect(res).toEqual([]);
  });

  it('returns single consumer entry for single valid block', async () => {
    const yaml = `
consumers:
  - name: frontend
    provider_repo: org/backend
    schema_type: openapi
    provider_spec_path: api/openapi.yaml
    provider_branch: dev
    `;
    const res = await parseConsumersFromYaml(yaml);
    expect(res).toEqual([{
      name: 'frontend',
      provider_repo: 'org/backend',
      schema_type: 'openapi',
      provider_spec_path: 'api/openapi.yaml',
      provider_branch: 'dev'
    }]);
  });

  it('defaults provider_branch to main if not present', async () => {
    const yaml = `
consumers:
  - name: frontend
    provider_repo: org/backend
    schema_type: openapi
    provider_spec_path: api/openapi.yaml
    `;
    const res = await parseConsumersFromYaml(yaml);
    expect(res).toEqual([{
      name: 'frontend',
      provider_repo: 'org/backend',
      schema_type: 'openapi',
      provider_spec_path: 'api/openapi.yaml',
      provider_branch: 'main'
    }]);
  });

  it('parses required_notice_days', async () => {
    const yaml = `
consumers:
  - name: my-consumer
    provider_repo: acme/api
    schema_type: openapi
    provider_spec_path: openapi.yaml
    required_notice_days: 30
    `;
    const res = await parseConsumersFromYaml(yaml);
    expect(res).toEqual([{
      name: 'my-consumer',
      provider_repo: 'acme/api',
      schema_type: 'openapi',
      provider_spec_path: 'openapi.yaml',
      provider_branch: 'main',
      required_notice_days: 30
    }]);
  });

  it('returns two entries when two valid consumers present', async () => {
    const yaml = `
consumers:
  - name: app1
    provider_repo: org/backend
    schema_type: openapi
    provider_spec_path: api/openapi.yaml
  - name: app2
    provider_repo: org/graphql
    schema_type: graphql
    provider_spec_path: schema.gql
    `;
    const res = await parseConsumersFromYaml(yaml);
    expect(res).toEqual([{
      name: 'app1',
      provider_repo: 'org/backend',
      schema_type: 'openapi',
      provider_spec_path: 'api/openapi.yaml',
      provider_branch: 'main'
    }, {
      name: 'app2',
      provider_repo: 'org/graphql',
      schema_type: 'graphql',
      provider_spec_path: 'schema.gql',
      provider_branch: 'main'
    }]);
  });
});

describe('syncToRegistry', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('resolves with synced count on success 200', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ synced: 2 })
    });
    globalThis.fetch = fetchMock;

    const res = await syncToRegistry('http://registry.api', 'token', {
      installation_id: 1, org: 'o', consumer_repo: 'o/c', consumer_github_repo_id: 2, commit_sha: 'a', dependencies: []
    });

    expect(fetchMock).toHaveBeenCalledWith('http://registry.api/api/v1/sync', expect.any(Object));
    expect(res).toEqual({ synced: 2 });
  });

  it('throws an Error on non-2xx response', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
      text: async () => 'internal server error'
    });
    globalThis.fetch = fetchMock;

    await expect(syncToRegistry('http://registry.api', 'token', {
      installation_id: 1, org: 'o', consumer_repo: 'o/c', consumer_github_repo_id: 2, commit_sha: 'a', dependencies: []
    })).rejects.toThrow('Failed to sync to registry: 500 internal server error');
  });
});



describe('crossRepoCheck', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  const dummyPayload = {
    installation_id: 1,
    org: 'myorg',
    provider_repo: 'myorg/backend-api',
    head_schema_content: 'openapi: 3.0.0',
    schema_type: 'openapi'
  };

  it('success — returns full ConsumerCompatibilityMatrix', async () => {
    const mockResponse = {
      total_consumers: 2,
      broken_consumers: 1,
      is_safe: false,
      results: [
        {
          consumer_repo: 'myorg/frontend',
          status: 'breaking',
          diff_report: { breaking: [{}], warning: [], info: [], summary: { breaking_count: 1, warning_count: 0, info_count: 0 } }
        },
        {
          consumer_repo: 'myorg/mobile-app',
          status: 'safe',
          diff_report: { breaking: [], warning: [], info: [], summary: { breaking_count: 0, warning_count: 0, info_count: 0 } }
        }
      ]
    };
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockResponse
    });
    globalThis.fetch = fetchMock;

    const res = await crossRepoCheck('http://registry.api', 'token', dummyPayload);
    expect(res).toEqual(mockResponse);
    expect(fetchMock).toHaveBeenCalledWith('http://registry.api/api/v1/cross-repo-check', expect.any(Object));
  });

  it('non-2xx response — returns safe default, does not throw', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
      text: async () => 'Internal Server Error'
    });
    globalThis.fetch = fetchMock;

    const res = await crossRepoCheck('http://registry.api', 'token', dummyPayload);
    expect(res).toEqual({
      total_consumers: 0,
      broken_consumers: 0,
      is_safe: true,
      results: []
    });
    expect(fetchMock).toHaveBeenCalledWith('http://registry.api/api/v1/cross-repo-check', expect.any(Object));
  });

  it('no consumers registered — returns safe default', async () => {
    const mockResponse = {
      total_consumers: 0,
      broken_consumers: 0,
      is_safe: true,
      results: []
    };
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockResponse
    });
    globalThis.fetch = fetchMock;

    const res = await crossRepoCheck('http://registry.api', 'token', dummyPayload);
    expect(res).toEqual(mockResponse);
    expect(fetchMock).toHaveBeenCalledWith('http://registry.api/api/v1/cross-repo-check', expect.any(Object));
  });

  it('fetch error (network issue) — returns safe default, does not throw', async () => {
    const fetchMock = vi.fn().mockRejectedValue(new Error('Network error'));
    globalThis.fetch = fetchMock;

    const res = await crossRepoCheck('http://registry.api', 'token', dummyPayload);
    expect(res).toEqual({
      total_consumers: 0,
      broken_consumers: 0,
      is_safe: true,
      results: []
    });
    expect(fetchMock).toHaveBeenCalledWith('http://registry.api/api/v1/cross-repo-check', expect.any(Object));
  });
});
