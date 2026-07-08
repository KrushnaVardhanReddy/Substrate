import { Env, DiffReport } from './types.js';
import { validateWebhookSignature, parsePREvent } from './webhook.js';
import { generateInstallationToken, fetchFileContent, postPRComment, setCommitStatus } from './github-client.js';
import { formatPRComment, formatMissingConfigComment, getCommitStatusState, getCommitStatusDescription } from './formatter.js';

// YAML parser mock/regex for the stub phase
function parseYaml(yaml: string): any {
  const result: any = {};
  const baseMatch = yaml.match(/base_schema:\s*(.+)/);
  if (baseMatch) result.base_schema = baseMatch[1].trim();
  const headMatch = yaml.match(/head_schema:\s*(.+)/);
  if (headMatch) result.head_schema = headMatch[1].trim();
  const onBreakingMatch = yaml.match(/on_breaking_change:\s*(.+)/);
  if (onBreakingMatch) result.on_breaking_change = onBreakingMatch[1].trim();
  const schemaTypeMatch = yaml.match(/schema_type:\s*(.+)/);
  if (schemaTypeMatch) result.schema_type = schemaTypeMatch[1].trim();
  return result;
}

async function callContainerService(containerUrl: string, baseSchema: string, headSchema: string, configContent: string, schemaType: string): Promise<DiffReport> {
  let response;
  try {
    response = await fetch(`${containerUrl}/diff`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        base_schema: baseSchema,
        head_schema: headSchema,
        config: configContent,
        schema_type: schemaType
      })
    });
  } catch (err) {
    throw new Error("container service error");
  }
  if (!response.ok) {
    throw new Error("container service error");
  }
  return await response.json();
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    // Step 1: Only accept POST
    if (request.method !== 'POST') {
      return new Response('Method Not Allowed', { status: 405 });
    }

    const signature = request.headers.get('X-Hub-Signature-256');
    if (!signature) {
      return new Response('Unauthorized', { status: 401 });
    }

    const body = await request.text();

    // Step 2: Validate HMAC
    const isValid = await validateWebhookSignature(signature, body, env.GITHUB_WEBHOOK_SECRET);
    if (!isValid) {
      return new Response('Unauthorized', { status: 401 });
    }

    // Step 3: Parse PR event
    const event = parsePREvent(request.headers, body);
    if (!event) {
      // Not a handled PR action, return 200 immediately
      return new Response('Ignored', { status: 200 });
    }

    try {
      // Step 4: Generate installation token
      const token = await generateInstallationToken(
        env.GITHUB_APP_ID,
        env.GITHUB_APP_PRIVATE_KEY,
        event.installationId
      );

      // Step 5: Set 'pending' commit status immediately
      await setCommitStatus(
        token,
        event.owner,
        event.repo,
        event.headSha,
        'pending',
        'Substrate is checking...'
      );

      // Step 6: Fetch substrate.yaml from head commit
      const configContent = await fetchFileContent(
        token,
        event.owner,
        event.repo,
        'substrate.yaml',
        event.headSha
      );

      // Step 7: If no substrate.yaml -> post missing-config comment, set 'pending' status, return 200
      if (!configContent) {
        const missingConfigMsg = formatMissingConfigComment();
        await postPRComment(token, event.owner, event.repo, event.prNumber, missingConfigMsg);

        // Ensure pending status remains (set neutrally above, spec says neutral/pending)
        await setCommitStatus(
          token,
          event.owner,
          event.repo,
          event.headSha,
          'pending',
          'substrate.yaml not found — run `substrate init` to activate'
        );

        return new Response('Missing Config', { status: 200 });
      }

      // Step 8: Parse config, fetch base + head spec files from GitHub API
      const config = parseYaml(configContent);
      let baseContent: string | null = null;
      let headContent: string | null = null;

      if (config.base_schema && config.head_schema) {
        baseContent = await fetchFileContent(token, event.owner, event.repo, config.base_schema, event.baseBranch);
        headContent = await fetchFileContent(token, event.owner, event.repo, config.head_schema, event.headSha);
      }

      if (!config.base_schema || !config.head_schema || baseContent === null || headContent === null) {
        const configErrorMsg = "## ⚠️ Substrate — Config Error\n\nCould not fetch spec files. Check `base_schema` and `head_schema` in your `substrate.yaml`.\n\n*Powered by [Substrate](https://github.com/KrushnaVardhanReddy/Substrate)*";
        await postPRComment(token, event.owner, event.repo, event.prNumber, configErrorMsg);
        await setCommitStatus(token, event.owner, event.repo, event.headSha, 'failure', 'Substrate config error — check substrate.yaml');
        return new Response('Config Error', { status: 200 });
      }

      // Step 9: REAL Container Service call
      let diffReport: DiffReport;
      try {
        const schemaType = config.schema_type || 'openapi';
        diffReport = await callContainerService(
          env.CONTAINER_SERVICE_URL,
          baseContent,
          headContent,
          configContent,
          schemaType
        );
      } catch (e: any) {
        await setCommitStatus(token, event.owner, event.repo, event.headSha, 'failure', 'Substrate engine error — retry later');
        return new Response('Engine Error', { status: 200 });
      }

      // Step 10: Post PR comment
      const commentBody = formatPRComment(diffReport, config);
      await postPRComment(token, event.owner, event.repo, event.prNumber, commentBody);

      // Step 11: Set final commit status
      const statusState = getCommitStatusState(diffReport, config);
      const statusDescription = getCommitStatusDescription(diffReport);
      await setCommitStatus(
        token,
        event.owner,
        event.repo,
        event.headSha,
        statusState,
        statusDescription
      );

      // Step 12: Return 200
      return new Response('OK', { status: 200 });

    } catch (e: any) {
      console.error(e);
      // Even on failure, webhook should return 200 to avoid GitHub disabling the app
      return new Response('Error Processing', { status: 200 });
    }
  }
};
