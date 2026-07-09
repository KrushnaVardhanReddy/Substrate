import { describe, it, expect, vi, beforeEach } from 'vitest';
import { parseConsumersFromYaml, syncToRegistry } from '../src/registry-client.js';

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
