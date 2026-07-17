import { StandardPREvent, StandardPushEvent } from '../../types';

export function parsePREvent(headers: Headers, body: string): StandardPREvent | null {
  const eventType = headers.get('X-Gitea-Event') || headers.get('X-Forgejo-Event');
  if (eventType !== 'pull_request') {
    return null;
  }

  try {
    const payload = JSON.parse(body);
    const action = payload.action;

    if (action !== 'opened' && action !== 'synchronized' && action !== 'reopened') {
      return null;
    }

    return {
      owner: payload.repository.owner.login || payload.repository.owner.username,
      repo: payload.repository.name,
      fullName: payload.repository.full_name,
      prNumber: payload.pull_request.number,
      headSha: payload.pull_request.head.sha,
      baseBranch: payload.pull_request.base.ref,
      installationId: 0 // Gitea does not use GitHub Apps installation IDs
    };
  } catch (e) {
    return null;
  }
}

export function parsePushEvent(headers: Headers, body: string): StandardPushEvent | null {
  const eventType = headers.get('X-Gitea-Event') || headers.get('X-Forgejo-Event');
  if (eventType !== 'push') {
    return null;
  }

  try {
    const payload = JSON.parse(body);

    if (payload.deleted === true) {
      return null;
    }

    if (payload.ref !== 'refs/heads/main' && payload.ref !== 'refs/heads/master') {
      return null;
    }

    return {
      ref: payload.ref,
      after: payload.after,
      installationId: 0,
      owner: payload.repository.owner.login || payload.repository.owner.username,
      repo: payload.repository.name,
      fullName: payload.repository.full_name,
      githubRepoId: payload.repository.id, // Using the Gitea repo ID
      installationOrgId: payload.repository.owner.id // Using the Gitea owner ID
    };
  } catch (e) {
    return null;
  }
}
