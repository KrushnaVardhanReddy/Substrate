import { Env, DiffReport, SyncDependency, CrossRepoCheckRequest, CrossRepoCheckResponse, AIAutofixRequest, AIAutofixResponse } from './types.js';
import { validateWebhookSignature, parsePREvent, parsePushEvent } from './webhook.js';
import { generateInstallationToken, fetchFileContent, postPRComment, setCommitStatus } from './github-client.js';
import { formatPRComment, formatMissingConfigComment, getCommitStatusState, getCommitStatusDescription, formatCrossRepoImpact } from './formatter.js';
import { parseConsumersFromYaml, syncToRegistry, crossRepoCheck } from './registry-client.js';


// YAML parser mock/regex for the stub phase
function parseYaml(yaml: string): any {
  const result: any = {};
  const baseMatch = yaml.match(/base_schema:\s*(.+)/);
  const headMatch = yaml.match(/head_schema:\s*(.+)/);
  const specMatch = yaml.match(/spec_path:\s*(.+)/);
  
  if (baseMatch) result.base_schema = baseMatch[1].trim();
  else if (specMatch) result.base_schema = specMatch[1].trim();
  
  if (headMatch) result.head_schema = headMatch[1].trim();
  else if (specMatch) result.head_schema = specMatch[1].trim();
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

    const eventType = request.headers.get('X-GitHub-Event');
    if (eventType === 'push') {
      const pushEvent = parsePushEvent(request.headers, body);
      if (!pushEvent) {
        return new Response('Ignored', { status: 200 });
      }

      try {
        const token = await generateInstallationToken(
          env.GITHUB_APP_ID,
          env.GITHUB_APP_PRIVATE_KEY,
          pushEvent.installationId
        );

        const configContent = await fetchFileContent(
          token,
          pushEvent.owner,
          pushEvent.repo,
          'substrate.yaml',
          pushEvent.after
        );

        if (!configContent) {
          return new Response('Ignored', { status: 200 });
        }

        const consumerEntries = await parseConsumersFromYaml(configContent);
        if (consumerEntries.length === 0) {
          return new Response('Ignored', { status: 200 });
        }

        const syncDependencies: SyncDependency[] = [];

        await Promise.all(consumerEntries.map(async (entry) => {
          try {
            // Fetch provider_github_repo_id using GitHub API
            const repoResponse = await fetch(`https://api.github.com/repos/${entry.provider_repo}`, {
              headers: {
                'Accept': 'application/vnd.github.v3+json',
                'Authorization': `Bearer ${token}`,
                'User-Agent': 'Substrate-GitHub-App'
              }
            });

            if (!repoResponse.ok) {
              console.error(`Failed to fetch repo ${entry.provider_repo}`);
              return;
            }

            const repoData = await repoResponse.json() as any;
            const providerOwner = repoData.owner.login;
            const providerRepoName = repoData.name;
            const providerGithubRepoId = repoData.id;

            const providerSpecContent = await fetchFileContent(
              token,
              providerOwner,
              providerRepoName,
              entry.provider_spec_path,
              entry.provider_branch
            );

            if (providerSpecContent) {
              syncDependencies.push({
                provider_repo: entry.provider_repo,
                provider_github_repo_id: providerGithubRepoId,
                schema_type: entry.schema_type,
                spec_path: entry.provider_spec_path,
                branch: entry.provider_branch,
                raw_content: providerSpecContent
              });
            }
          } catch (err) {
            console.error(`Error processing consumer entry ${entry.name}`, err);
          }
        }));

        const syncResult = await syncToRegistry(env.REGISTRY_API_URL, env.REGISTRY_API_TOKEN, {
          installation_id: pushEvent.installationId,
          org: pushEvent.owner,
          consumer_repo: pushEvent.fullName,
          consumer_github_repo_id: pushEvent.githubRepoId,
          commit_sha: pushEvent.after,
          dependencies: syncDependencies
        });

        return new Response(JSON.stringify(syncResult), { status: 200, headers: { 'Content-Type': 'application/json' } });
      } catch (e: any) {
        console.error(e);
        return new Response('Error Processing Push', { status: 200 });
      }
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
        console.error("Engine Call Failed:", e);
        await setCommitStatus(token, event.owner, event.repo, event.headSha, 'failure', 'Substrate engine error — retry later');
        return new Response('Engine Error', { status: 200 });
      }

      // Step 9.5: Cross Repo Check
      let crossRepoResponse: CrossRepoCheckResponse | undefined;
      const schemaType = config.schema_type || 'openapi';

      if (env.REGISTRY_API_URL) {
        const payload: CrossRepoCheckRequest = {
          installation_id: event.installationId,
          org: event.owner,
          provider_repo: event.fullName,
          head_schema_content: headContent,
          schema_type: schemaType,
          config_content: configContent
        };
        crossRepoResponse = await crossRepoCheck(env.REGISTRY_API_URL, env.REGISTRY_API_TOKEN, payload);
      } else {
        crossRepoResponse = { total_consumers: 0, broken_consumers: 0, is_safe: true, results: [] };
      }

      const crossRepoSection = formatCrossRepoImpact(crossRepoResponse);

      // Step 9.75: AI Autofix
      let aiExplanation: string | undefined;
      let aiSafePatch: string | undefined;

      if (diffReport.summary.breaking_count > 0 && env.REGISTRY_API_URL) {
        try {
          const autofixReq: AIAutofixRequest = {
            provider_repo: event.fullName,
            schema_type: config.schema_type || 'openapi',
            current_schema: baseContent,
            proposed_schema: headContent,
            breaking_changes: diffReport.breaking_changes
          };

          const autofixRes = await fetch(`${env.REGISTRY_API_URL}/api/v1/ai/autofix`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${env.REGISTRY_API_TOKEN}`
            },
            body: JSON.stringify(autofixReq)
          });

          if (autofixRes.ok) {
            const autofixData = await autofixRes.json() as AIAutofixResponse;
            aiExplanation = autofixData.explanation;
            aiSafePatch = autofixData.safe_patch;
          } else {
            console.error(`AI Autofix failed with status ${autofixRes.status}`);
          }
        } catch (e) {
          console.error("AI Autofix request failed:", e);
        }
      }

      // Step 10: Post PR comment
      let commentBody = formatPRComment(diffReport, config, env.DASHBOARD_URL, event.owner, event.repo, event.prNumber, aiExplanation, aiSafePatch);
      if (crossRepoSection) {
        commentBody += "\n" + crossRepoSection;
      }
      await postPRComment(token, event.owner, event.repo, event.prNumber, commentBody);

      // Step 11: Set final commit status
      const statusState = getCommitStatusState(diffReport, config, crossRepoResponse);
      const statusDescription = getCommitStatusDescription(diffReport, crossRepoResponse);
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
