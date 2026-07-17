import { VCSClient } from '../../types';

export class GiteaProvider implements VCSClient {
  private apiUrl: string;
  private token: string;

  constructor(apiUrl: string, token: string) {
    this.apiUrl = apiUrl.replace(/\/$/, '');
    this.token = token;
  }

  async fetchFileContent(owner: string, repo: string, path: string, ref: string): Promise<string | null> {
    const response = await fetch(
      `${this.apiUrl}/api/v1/repos/${owner}/${repo}/contents/${path}?ref=${ref}`,
      {
        headers: {
          'Authorization': `token ${this.token}`,
          'Accept': 'application/json'
        }
      }
    );

    if (response.status === 404) {
      return null;
    }

    if (!response.ok) {
      throw new Error(`Failed to fetch file content: ${response.status} ${await response.text()}`);
    }

    const data: any = await response.json();
    if (data.type !== 'file' || !data.content) {
      return null;
    }

    return atob(data.content);
  }

  async postPRComment(owner: string, repo: string, prNumber: number, body: string): Promise<void> {
    const getResponse = await fetch(
      `${this.apiUrl}/api/v1/repos/${owner}/${repo}/issues/${prNumber}/comments`,
      {
        headers: {
          'Authorization': `token ${this.token}`,
          'Accept': 'application/json'
        }
      }
    );

    if (!getResponse.ok) {
      throw new Error(`Failed to fetch comments: ${getResponse.status} ${await getResponse.text()}`);
    }

    const comments: any[] = await getResponse.json();

    const existingComment = comments.find(c =>
      c.body.startsWith('## ') && c.body.includes('Substrate')
    );

    if (existingComment) {
      const patchResponse = await fetch(
        `${this.apiUrl}/api/v1/repos/${owner}/${repo}/issues/comments/${existingComment.id}`,
        {
          method: 'PATCH',
          headers: {
            'Authorization': `token ${this.token}`,
            'Accept': 'application/json',
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ body })
        }
      );

      if (!patchResponse.ok) {
        throw new Error(`Failed to update comment: ${patchResponse.status} ${await patchResponse.text()}`);
      }
    } else {
      const postResponse = await fetch(
        `${this.apiUrl}/api/v1/repos/${owner}/${repo}/issues/${prNumber}/comments`,
        {
          method: 'POST',
          headers: {
            'Authorization': `token ${this.token}`,
            'Accept': 'application/json',
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ body })
        }
      );

      if (!postResponse.ok) {
        throw new Error(`Failed to post comment: ${postResponse.status} ${await postResponse.text()}`);
      }
    }
  }

  async setCommitStatus(owner: string, repo: string, sha: string, state: 'success' | 'failure' | 'pending', description: string): Promise<void> {
    const response = await fetch(
      `${this.apiUrl}/api/v1/repos/${owner}/${repo}/statuses/${sha}`,
      {
        method: 'POST',
        headers: {
          'Authorization': `token ${this.token}`,
          'Accept': 'application/json',
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          state,
          description,
          context: 'substrate/breaking-changes'
        })
      }
    );

    if (!response.ok) {
      throw new Error(`Failed to set commit status: ${response.status} ${await response.text()}`);
    }
  }

  async fetchPRFiles(owner: string, repo: string, prNumber: number): Promise<string[]> {
    const response = await fetch(
      `${this.apiUrl}/api/v1/repos/${owner}/${repo}/pulls/${prNumber}/files?limit=100`,
      {
        headers: {
          'Authorization': `token ${this.token}`,
          'Accept': 'application/json'
        }
      }
    );

    if (!response.ok) {
      console.error(`Failed to fetch PR files: ${response.status}`);
      return [];
    }

    const files: any[] = await response.json();
    return files.map(f => f.filename);
  }
}
