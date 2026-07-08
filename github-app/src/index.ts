import { Env, DiffReport } from './types';
import { validateWebhookSignature, parsePREvent } from './webhook';
import { generateInstallationToken, fetchFileContent, postPRComment, setCommitStatus } from './github-client';

// YAML parser mock/regex for the stub phase
function parseYaml(yaml: string): any {
  const result: any = {};
  const baseMatch = yaml.match(/base_schema:\s*(.+)/);
  if (baseMatch) result.base_schema = baseMatch[1].trim();
  const headMatch = yaml.match(/head_schema:\s*(.+)/);
  if (headMatch) result.head_schema = headMatch[1].trim();
  const onBreakingMatch = yaml.match(/on_breaking_change:\s*(.+)/);
  if (onBreakingMatch) result.on_breaking_change = onBreakingMatch[1].trim();
  return result;
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
        const missingConfigMsg = "## 👋 Substrate is installed but not configured\n\nRun `substrate init` to activate.\n\n*Powered by [Substrate](https://github.com/KrushnaVardhanReddy/Substrate)*";
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

      if (config.base_schema && config.head_schema) {
        // We fetch these just to fulfill step 8 (they will be sent to the Container Service in P2-T02)
        await fetchFileContent(token, event.owner, event.repo, config.base_schema, event.baseBranch);
        await fetchFileContent(token, event.owner, event.repo, config.head_schema, event.headSha);
      }

      // Step 9: STUB diff result
      const diffReport: DiffReport = {
        breaking: [],
        warning: [],
        info: [],
        summary: {
          breaking_count: 0,
          warning_count: 0,
          info_count: 0
        }
      };

      // Step 10: Post PR comment (placeholder)
      const allClearMsg = "## ✅ Substrate — All Clear\n\nNo breaking changes detected in this PR. Safe to merge. 🎉\n\n*Powered by [Substrate](https://github.com/KrushnaVardhanReddy/Substrate)*";
      await postPRComment(token, event.owner, event.repo, event.prNumber, allClearMsg);

      // Step 11: Set final commit status
      await setCommitStatus(
        token,
        event.owner,
        event.repo,
        event.headSha,
        'success',
        'All clear — no breaking changes'
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
