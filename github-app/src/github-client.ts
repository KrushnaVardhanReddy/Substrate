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
