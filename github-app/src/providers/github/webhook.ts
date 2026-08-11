// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { StandardPREvent, StandardPushEvent } from '../../types';

export function parsePREvent(headers: Headers, body: string): StandardPREvent | null {
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
      installationId: payload.installation?.id || 0
    };
  } catch (e) {
    return null;
  }
}

export function parsePushEvent(headers: Headers, body: string): StandardPushEvent | null {
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
