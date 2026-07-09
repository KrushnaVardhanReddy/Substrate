import type { CrossRepoCheckRequest, CrossRepoCheckResponse, ConsumerResult } from './types.js';
import { ConsumerEntry, SyncRequest } from './types.js';

export async function parseConsumersFromYaml(yamlContent: string): Promise<ConsumerEntry[]> {
  const consumers: ConsumerEntry[] = [];

  const consumersMatch = yamlContent.match(/consumers:\s*([\s\S]*)/);
  if (!consumersMatch) return [];

  // Very naive parser - finding blocks starting with `- name:`
  const consumerBlocks = consumersMatch[1].split(/(?=\n\s*-\s*name:)/).filter(b => b.trim() !== '');

  for (const block of consumerBlocks) {
    if (!block.trim().startsWith('- name:')) continue;

    const nameMatch = block.match(/-\s*name:\s*(.+)/);
    const providerRepoMatch = block.match(/provider_repo:\s*(.+)/);
    const schemaTypeMatch = block.match(/schema_type:\s*(.+)/);
    const providerSpecPathMatch = block.match(/provider_spec_path:\s*(.+)/);
    const providerBranchMatch = block.match(/provider_branch:\s*(.+)/);

    const stripQuotes = (val: string) => val.replace(/^["']|["']$/g, '').trim();

    if (nameMatch && providerRepoMatch && schemaTypeMatch && providerSpecPathMatch) {
      consumers.push({
        name: stripQuotes(nameMatch[1]),
        provider_repo: stripQuotes(providerRepoMatch[1]),
        schema_type: stripQuotes(schemaTypeMatch[1]),
        provider_spec_path: stripQuotes(providerSpecPathMatch[1]),
        provider_branch: providerBranchMatch ? stripQuotes(providerBranchMatch[1]) : 'main'
      });
    }
  }

  return consumers;
}

export async function syncToRegistry(
  registryUrl: string,
  token: string,
  payload: SyncRequest
): Promise<{ synced: number }> {
  const url = `${registryUrl.replace(/\/$/, '')}/api/v1/sync`;

  const response = await fetch(url, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`Failed to sync to registry: ${response.status} ${errorText}`);
  }

  const result = await response.json() as { synced: number };
  return result;
}



export async function crossRepoCheck(
  registryUrl: string,
  token: string,
  payload: CrossRepoCheckRequest
): Promise<CrossRepoCheckResponse> {
  const safeDefault: CrossRepoCheckResponse = {
    total_consumers: 0,
    broken_consumers: 0,
    is_safe: true,
    results: []
  };

  try {
    const url = `${registryUrl.replace(/\/$/, '')}/api/v1/cross-repo-check`;
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(payload)
    });

    if (!response.ok) {
      console.error(`Failed cross-repo check: ${response.status}`);
      return safeDefault;
    }

    const result = await response.json() as CrossRepoCheckResponse;
    if (!result.results) {
        return safeDefault;
    }
    return result;
  } catch (err) {
    console.error('Error during cross-repo check', err);
    return safeDefault;
  }
}
