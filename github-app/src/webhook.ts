import { PushEvent } from './types.js';

export interface PREvent {
  owner: string;
  repo: string;
  fullName: string;
  prNumber: number;
  headSha: string;
  baseBranch: string;
  installationId: number;
}

function hexToArrayBuffer(hex: string): ArrayBuffer {
  const bytes = new Uint8Array(Math.ceil(hex.length / 2));
  for (let i = 0; i < bytes.length; i++) {
    bytes[i] = parseInt(hex.substring(i * 2, i * 2 + 2), 16);
  }
  return bytes.buffer;
}

export async function validateWebhookSignature(
  signature: string,
  body: string,
  secret: string
): Promise<boolean> {
  if (!signature.startsWith('sha256=')) {
    return false;
  }

  const expectedSigHex = signature.slice(7);

  const encoder = new TextEncoder();

  const key = await crypto.subtle.importKey(
    'raw',
    encoder.encode(secret),
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign', 'verify']
  );

  // Create HMAC of the body
  const bodyBuffer = encoder.encode(body);

  try {
    const signatureBuffer = hexToArrayBuffer(expectedSigHex);
    // Use verify for constant-time comparison
    return await crypto.subtle.verify(
      'HMAC',
      key,
      signatureBuffer,
      bodyBuffer
    );
  } catch (e) {
    return false;
  }
}

export function parsePREvent(headers: Headers, body: string): PREvent | null {
  const eventType = headers.get('X-GitHub-Event');
  if (eventType !== 'pull_request') {
    return null;
  }

  try {
    const payload = JSON.parse(body);
    const action = payload.action;

    if (action !== 'opened' && action !== 'synchronize' && action !== 'reopened') {
      return null;
    }

    return {
      owner: payload.repository.owner.login,
      repo: payload.repository.name,
      fullName: payload.repository.full_name,
      prNumber: payload.pull_request.number,
      headSha: payload.pull_request.head.sha,
      baseBranch: payload.pull_request.base.ref,
      installationId: payload.installation.id
    };
  } catch (e) {
    return null;
  }
}

export function parsePushEvent(headers: Headers, body: string): PushEvent | null {
  const eventType = headers.get('X-GitHub-Event');
  if (eventType !== 'push') {
    return null;
  }

  try {
    const payload = JSON.parse(body);

    if (payload.deleted === true) {
      return null;
    }

    if (payload.ref !== 'refs/heads/main') {
      return null;
    }

    return {
      ref: payload.ref,
      after: payload.after,
      installationId: payload.installation?.id || 0,
      owner: payload.repository.owner.login || payload.repository.owner.name,
      repo: payload.repository.name,
      fullName: payload.repository.full_name,
      githubRepoId: payload.repository.id,
      installationOrgId: payload.repository.owner.id
    };
  } catch (e) {
    return null;
  }
}

export function parseInstallationRepositoriesEvent(headers: Headers, body: string): import('./types.js').InstallationRepositoriesEvent | null {
  const eventType = headers.get('X-GitHub-Event');
  if (eventType !== 'installation_repositories') {
    return null;
  }

  try {
    const payload = JSON.parse(body);
    if (!payload.action || !payload.installation) {
      return null;
    }
    return payload as import('./types.js').InstallationRepositoriesEvent;
  } catch (e) {
    return null;
  }
}

export function parseInstallationEvent(headers: Headers, body: string): import('./types.js').InstallationEvent | null {
  const eventType = headers.get('X-GitHub-Event');
  if (eventType !== 'installation') {
    return null;
  }

  try {
    const payload = JSON.parse(body);
    if (!payload.action || !payload.installation) {
      return null;
    }
    return payload as import('./types.js').InstallationEvent;
  } catch (e) {
    return null;
  }
}
