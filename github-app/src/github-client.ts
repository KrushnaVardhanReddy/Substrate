function base64UrlEncode(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer);
  let binary = '';
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary)
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '');
}

function strToUint8Array(str: string): Uint8Array {
  const buf = new Uint8Array(str.length);
  for (let i = 0; i < str.length; i++) {
    buf[i] = str.charCodeAt(i);
  }
  return buf;
}

function parsePEM(pem: string): ArrayBuffer {
  const b64 = pem
    .replace(/-----[A-Z ]+-----/g, '')
    .replace(/\s+/g, '');
  const binary = atob(b64);
  return strToUint8Array(binary).buffer as ArrayBuffer;
}

export async function generateInstallationToken(
  appId: string,
  privateKey: string,
  installationId: number
): Promise<string> {
  const header = { alg: 'RS256', typ: 'JWT' };
  const now = Math.floor(Date.now() / 1000);
  const payload = {
    iss: appId,
    iat: now - 60,
    exp: now + 540
  };

  const encoder = new TextEncoder();
  const headerEncoded = base64UrlEncode(encoder.encode(JSON.stringify(header)).buffer);
  const payloadEncoded = base64UrlEncode(encoder.encode(JSON.stringify(payload)).buffer);
  const dataToSign = `${headerEncoded}.${payloadEncoded}`;

  const keyBuffer = parsePEM(privateKey);

  const key = await crypto.subtle.importKey(
    'pkcs8',
    keyBuffer,
    {
      name: 'RSASSA-PKCS1-v1_5',
      hash: 'SHA-256'
    },
    false,
    ['sign']
  );

  const signature = await crypto.subtle.sign(
    'RSASSA-PKCS1-v1_5',
    key,
    encoder.encode(dataToSign)
  );

  const signatureEncoded = base64UrlEncode(signature);
  const jwt = `${dataToSign}.${signatureEncoded}`;

  const response = await fetch(
    `https://api.github.com/app/installations/${installationId}/access_tokens`,
    {
      method: 'POST',
      headers: {
        'Accept': 'application/vnd.github.v3+json',
        'Authorization': `Bearer ${jwt}`,
        'User-Agent': 'Substrate-GitHub-App'
      }
    }
  );

  if (!response.ok) {
    throw new Error(`Failed to generate installation token: ${response.status} ${await response.text()}`);
  }

  const data: any = await response.json();
  return data.token;
}

export async function fetchFileContent(
  token: string,
  owner: string,
  repo: string,
  path: string,
  ref: string
): Promise<string | null> {
  console.log(`fetchFileContent: token starts with ${token.substring(0, 8)}`);
  const response = await fetch(
    `https://api.github.com/repos/${owner}/${repo}/contents/${path}?ref=${ref}`,
    {
      headers: {
        'Accept': 'application/vnd.github.v3+json',
        'Authorization': `Bearer ${token}`,
        'User-Agent': 'Substrate-GitHub-App'
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

  // Base64 decoding
  return atob(data.content);
}

export async function postPRComment(
  token: string,
  owner: string,
  repo: string,
  prNumber: number,
  body: string
): Promise<void> {
  // First, check for existing comment
  const getResponse = await fetch(
    `https://api.github.com/repos/${owner}/${repo}/issues/${prNumber}/comments`,
    {
      headers: {
        'Accept': 'application/vnd.github.v3+json',
        'Authorization': `Bearer ${token}`,
        'User-Agent': 'Substrate-GitHub-App'
      }
    }
  );

  if (!getResponse.ok) {
    throw new Error(`Failed to fetch comments: ${getResponse.status} ${await getResponse.text()}`);
  }

  const comments: any[] = await getResponse.json();

  // Find bot comment starting with Substrate header
  const existingComment = comments.find(c =>
    c.user.type === 'Bot' && c.body.startsWith('## ') && c.body.includes('Substrate')
  );

  if (existingComment) {
    // PATCH existing comment
    const patchResponse = await fetch(
      existingComment.url,
      {
        method: 'PATCH',
        headers: {
          'Accept': 'application/vnd.github.v3+json',
          'Authorization': `Bearer ${token}`,
          'User-Agent': 'Substrate-GitHub-App',
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ body })
      }
    );

    if (!patchResponse.ok) {
      throw new Error(`Failed to update comment: ${patchResponse.status} ${await patchResponse.text()}`);
    }
  } else {
    // POST new comment
    const postResponse = await fetch(
      `https://api.github.com/repos/${owner}/${repo}/issues/${prNumber}/comments`,
      {
        method: 'POST',
        headers: {
          'Accept': 'application/vnd.github.v3+json',
          'Authorization': `Bearer ${token}`,
          'User-Agent': 'Substrate-GitHub-App',
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

export async function setCommitStatus(
  token: string,
  owner: string,
  repo: string,
  sha: string,
  state: 'success' | 'failure' | 'pending',
  description: string
): Promise<void> {
  const response = await fetch(
    `https://api.github.com/repos/${owner}/${repo}/statuses/${sha}`,
    {
      method: 'POST',
      headers: {
        'Accept': 'application/vnd.github.v3+json',
        'Authorization': `Bearer ${token}`,
        'User-Agent': 'Substrate-GitHub-App',
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

export async function checkFileExists(
  token: string,
  owner: string,
  repo: string,
  path: string,
  branch: string = 'main'
): Promise<boolean> {
  const response = await fetch(
    `https://api.github.com/repos/${owner}/${repo}/contents/${path}?ref=${branch}`,
    {
      headers: {
        'Accept': 'application/vnd.github.v3+json',
        'Authorization': `Bearer ${token}`,
        'User-Agent': 'Substrate-GitHub-App'
      }
    }
  );

  return response.status === 200;
}

export async function getDefaultBranch(
  token: string,
  owner: string,
  repo: string
): Promise<string> {
  const response = await fetch(
    `https://api.github.com/repos/${owner}/${repo}`,
    {
      headers: {
        'Accept': 'application/vnd.github.v3+json',
        'Authorization': `Bearer ${token}`,
        'User-Agent': 'Substrate-GitHub-App'
      }
    }
  );
  if (!response.ok) {
    throw new Error(`Failed to fetch repo info: ${response.status} ${await response.text()}`);
  }
  const data: any = await response.json();
  return data.default_branch;
}

export async function getBranchSha(
  token: string,
  owner: string,
  repo: string,
  branch: string
): Promise<string> {
  const response = await fetch(
    `https://api.github.com/repos/${owner}/${repo}/git/ref/heads/${branch}`,
    {
      headers: {
        'Accept': 'application/vnd.github.v3+json',
        'Authorization': `Bearer ${token}`,
        'User-Agent': 'Substrate-GitHub-App'
      }
    }
  );
  if (!response.ok) {
    throw new Error(`Failed to fetch ref info: ${response.status} ${await response.text()}`);
  }
  const data: any = await response.json();
  return data.object.sha;
}

export async function createBranch(
  token: string,
  owner: string,
  repo: string,
  branchName: string,
  sha: string
): Promise<void> {
  const response = await fetch(
    `https://api.github.com/repos/${owner}/${repo}/git/refs`,
    {
      method: 'POST',
      headers: {
        'Accept': 'application/vnd.github.v3+json',
        'Authorization': `Bearer ${token}`,
        'User-Agent': 'Substrate-GitHub-App',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        ref: `refs/heads/${branchName}`,
        sha: sha
      })
    }
  );

  if (!response.ok) {
    // 422 Unprocessable Entity can mean reference already exists, handle gracefully if needed
    if (response.status !== 422) {
        throw new Error(`Failed to create branch: ${response.status} ${await response.text()}`);
    }
  }
}

export async function createCommitWithFile(
  token: string,
  owner: string,
  repo: string,
  branch: string,
  path: string,
  content: string,
  message: string
): Promise<void> {
  // Check if file exists first to get sha for update
  let sha: string | undefined;
  const getResp = await fetch(
    `https://api.github.com/repos/${owner}/${repo}/contents/${path}?ref=${branch}`,
    {
      headers: {
        'Accept': 'application/vnd.github.v3+json',
        'Authorization': `Bearer ${token}`,
        'User-Agent': 'Substrate-GitHub-App'
      }
    }
  );
  if (getResp.ok) {
    const data: any = await getResp.json();
    sha = data.sha;
  }

  const response = await fetch(
    `https://api.github.com/repos/${owner}/${repo}/contents/${path}`,
    {
      method: 'PUT',
      headers: {
        'Accept': 'application/vnd.github.v3+json',
        'Authorization': `Bearer ${token}`,
        'User-Agent': 'Substrate-GitHub-App',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        message,
        content: btoa(content),
        branch,
        ...(sha ? { sha } : {})
      })
    }
  );
  if (!response.ok) {
    throw new Error(`Failed to create commit: ${response.status} ${await response.text()}`);
  }
}

export async function createPullRequest(
  token: string,
  owner: string,
  repo: string,
  title: string,
  head: string,
  base: string,
  body: string
): Promise<number | null> {
  const response = await fetch(
    `https://api.github.com/repos/${owner}/${repo}/pulls`,
    {
      method: 'POST',
      headers: {
        'Accept': 'application/vnd.github.v3+json',
        'Authorization': `Bearer ${token}`,
        'User-Agent': 'Substrate-GitHub-App',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        title,
        head,
        base,
        body
      })
    }
  );
  if (!response.ok) {
    if (response.status === 422) { // PR might already exist
        return null;
    }
    throw new Error(`Failed to create pull request: ${response.status} ${await response.text()}`);
  }
  const data: any = await response.json();
  return data.number;
}

export async function fetchPRFiles(
  token: string,
  owner: string,
  repo: string,
  prNumber: number
): Promise<string[]> {
  const response = await fetch(
    `https://api.github.com/repos/${owner}/${repo}/pulls/${prNumber}/files?per_page=100`,
    {
      headers: {
        'Accept': 'application/vnd.github.v3+json',
        'Authorization': `Bearer ${token}`,
        'User-Agent': 'Substrate-GitHub-App'
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
export async function createIssue(
  token: string,
  owner: string,
  repo: string,
  title: string,
  body: string
): Promise<any> {
  const url = `https://api.github.com/repos/${owner}/${repo}/issues`;
  const res = await fetch(url, {
    method: 'POST',
    headers: {
      'Authorization': `token ${token}`,
      'Accept': 'application/vnd.github.v3+json',
      'Content-Type': 'application/json',
      'User-Agent': 'Substrate-GitHub-App',
    },
    body: JSON.stringify({ title, body }),
  });

  if (!res.ok) {
    throw new Error(`Failed to create issue: ${res.statusText}`);
  }

  return res.json();
}

export async function findOpenIssue(
  token: string,
  owner: string,
  repo: string,
  titlePrefix: string
): Promise<any> {
  const query = encodeURIComponent(`repo:${owner}/${repo} is:issue is:open in:title "${titlePrefix}"`);
  const url = `https://api.github.com/search/issues?q=${query}`;
  const res = await fetch(url, {
    method: 'GET',
    headers: {
      'Authorization': `token ${token}`,
      'Accept': 'application/vnd.github.v3+json',
      'User-Agent': 'Substrate-GitHub-App',
    },
  });

  if (!res.ok) {
    throw new Error(`Failed to search issues: ${res.statusText}`);
  }

  return res.json();
}

export async function closeIssue(
  token: string,
  owner: string,
  repo: string,
  issueNumber: number,
  commentBody: string
): Promise<void> {
  // Post the closing comment
  await postPRComment(token, owner, repo, issueNumber, commentBody);

  // Close the issue
  const url = `https://api.github.com/repos/${owner}/${repo}/issues/${issueNumber}`;
  const res = await fetch(url, {
    method: 'PATCH',
    headers: {
      'Authorization': `token ${token}`,
      'Accept': 'application/vnd.github.v3+json',
      'Content-Type': 'application/json',
      'User-Agent': 'Substrate-GitHub-App',
    },
    body: JSON.stringify({ state: 'closed' }),
  });

  if (!res.ok) {
    throw new Error(`Failed to close issue: ${res.statusText}`);
  }
}
