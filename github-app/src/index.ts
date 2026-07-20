import { Env, DiffReport, SyncDependency, CrossRepoCheckRequest, CrossRepoCheckResponse, AIAutofixRequest, AIAutofixResponse, VCSClient } from './types.js';
import { validateWebhookSignature, parseInstallationRepositoriesEvent, parseInstallationEvent } from './webhook.js';
import { generateInstallationToken } from './github-client.js';
import { formatPRComment, formatMissingConfigComment, getCommitStatusState, getCommitStatusDescription, formatCrossRepoImpact } from './formatter.js';
import { parseConsumersFromYaml, syncToRegistry, crossRepoCheck } from './registry-client.js';
import { processAutoDiscovery } from './discovery.js';
import { parseGitHubPREvent, parseGitHubPushEvent, GitHubProvider } from './providers/github/index.js';
import { parseGiteaPREvent, parseGiteaPushEvent, GiteaProvider } from './providers/gitea/index.js';


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

async function getVCSProvider(request: Request, env: Env, eventData: any): Promise<{provider: VCSClient, eventOwner: string, eventRepo: string} | null> {
  const isGitea = request.headers.has('X-Gitea-Event') || request.headers.has('X-Forgejo-Event');

  if (isGitea && env.GITEA_API_URL && env.GITEA_TOKEN) {
    return {
      provider: new GiteaProvider(env.GITEA_API_URL, env.GITEA_TOKEN),
      eventOwner: eventData.owner,
      eventRepo: eventData.repo
    };
  } else if (!isGitea) {
    if (!eventData.installationId) return null;
    const token = await generateInstallationToken(env.GITHUB_APP_ID, env.GITHUB_APP_PRIVATE_KEY, eventData.installationId);
    return {
      provider: new GitHubProvider(token),
      eventOwner: eventData.owner,
      eventRepo: eventData.repo
    };
  }
  return null;
}

export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    // Step 1: Only accept POST
    if (request.method !== 'POST') {
      return new Response('Method Not Allowed', { status: 405 });
    }

    const isGitea = request.headers.has('X-Gitea-Event') || request.headers.has('X-Forgejo-Event');
    const isGitHub = request.headers.has('X-GitHub-Event');

    if (!isGitea && !isGitHub) {
      return new Response('Unknown Webhook Source', { status: 400 });
    }

    const body = await request.text();

    if (isGitHub && !isGitea) {
      const signature = request.headers.get('X-Hub-Signature-256');
      if (!signature) {
        return new Response('Unauthorized', { status: 401 });
      }

      const isValid = await validateWebhookSignature(signature, body, env.GITHUB_WEBHOOK_SECRET);
      if (!isValid) {
        return new Response('Unauthorized', { status: 401 });
      }

      const eventType = request.headers.get('X-GitHub-Event');

      if (eventType === 'installation_repositories') {
        const installReposEvent = parseInstallationRepositoriesEvent(request.headers, body);
        if (installReposEvent && installReposEvent.action === 'added' && installReposEvent.repositories_added) {
          ctx.waitUntil(processAutoDiscovery(env, installReposEvent.installation.id, installReposEvent.repositories_added).catch(console.error));
        }
        return new Response('Accepted', { status: 202 });
      }

      if (eventType === 'installation') {
        const installEvent = parseInstallationEvent(request.headers, body);
        if (installEvent && installEvent.action === 'created' && installEvent.repositories) {
           ctx.waitUntil(processAutoDiscovery(env, installEvent.installation.id, installEvent.repositories).catch(console.error));
        }
        return new Response('Accepted', { status: 202 });
      }
    }

    const pushEvent = isGitea ? parseGiteaPushEvent(request.headers, body) : parseGitHubPushEvent(request.headers, body);
    const prEvent = isGitea ? parseGiteaPREvent(request.headers, body) : parseGitHubPREvent(request.headers, body);

    let payloadStr = body;
    let prAction = '';
    let eventType = request.headers.get('X-GitHub-Event');
    try {
      const parsed = JSON.parse(payloadStr);
      prAction = parsed.action;
    } catch(e) {}

    // Handle PR Close specifically for expiring preview tokens
    if (eventType === 'pull_request' && (prAction === 'closed' || prAction === 'merged') && prEvent) {
      if (env.REGISTRY_API_URL) {
        try {
          await fetch(`${env.REGISTRY_API_URL}/api/v1/preview/expire-pr`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${env.REGISTRY_API_TOKEN}`
            },
            body: JSON.stringify({
              org: prEvent.owner,
              repo: prEvent.repo,
              pr_number: prEvent.prNumber
            })
          });
        } catch (e) {
          console.error("Failed to expire preview session:", e);
        }
      }
      return new Response('PR Closed/Merged Processed', { status: 200 });
    }

    if (!pushEvent && !prEvent) {
      console.log('Ignored: No valid push or PR event parsed');
      return new Response('Ignored', { status: 200 });
    }

    const isPushEvent = !!pushEvent;
    const eventData = pushEvent || prEvent;

    if (!eventData) {
        return new Response('No Event Data', { status: 400 });
    }

    let providerResult;
    try {
        providerResult = await getVCSProvider(request, env, eventData);
    } catch (e) {
        console.error("Token Generation Failed:", e);
        return new Response('Auth Error', { status: 200 });
    }

    if (!providerResult) {
         console.log('Auth Error: providerResult is null. Check env vars GITEA_API_URL/GITEA_TOKEN.');
         return new Response('Auth Error', { status: 200 });
    }

    const { provider, eventOwner, eventRepo } = providerResult;

    if (isPushEvent) {
      try {
        const pushEv = pushEvent!;
        const configContent = await provider.fetchFileContent(
          eventOwner,
          eventRepo,
          'substrate.yaml',
          pushEv.after
        );

        if (!configContent) {
          console.log(`Ignored: No substrate.yaml found in ${eventOwner}/${eventRepo} at ${pushEv.after}`);
          return new Response('Ignored', { status: 200 });
        }

        const consumerEntries = await parseConsumersFromYaml(configContent);
        if (consumerEntries.length === 0) {
          console.log('Ignored: No consumers found in substrate.yaml');
          return new Response('Ignored', { status: 200 });
        }

        let syncedTotal = 0;

        const hashString = (str: string): number => {
          let hash = 0;
          for (let i = 0; i < str.length; i++) {
            hash = Math.imul(31, hash) + str.charCodeAt(i) | 0;
          }
          return Math.abs(hash);
        };

        await Promise.all(consumerEntries.map(async (entry) => {
          try {
            const providerOwner = entry.provider_repo.split('/')[0] || eventOwner;
            const providerRepoName = entry.provider_repo.split('/')[1] || entry.provider_repo;

            const providerSpecContent = await provider.fetchFileContent(
              providerOwner,
              providerRepoName,
              entry.provider_spec_path,
              entry.provider_branch
            );

            if (providerSpecContent) {
              const consumerFullName = `${eventOwner}/${entry.name}`;
              const syncResult = await syncToRegistry(env.REGISTRY_API_URL, env.REGISTRY_API_TOKEN, {
                installation_id: pushEv.installationId,
                org: eventOwner,
                consumer_repo: consumerFullName,
                consumer_github_repo_id: hashString(consumerFullName),
                commit_sha: pushEv.after,
                dependencies: [{
                  provider_repo: entry.provider_repo,
                  provider_github_repo_id: hashString(entry.provider_repo), // Using hash as fallback for repo ID
                  schema_type: entry.schema_type,
                  spec_path: entry.provider_spec_path,
                  branch: entry.provider_branch,
                  raw_content: providerSpecContent,
                  required_notice_days: entry.required_notice_days
                }]
              });
              if (syncResult && syncResult.synced !== undefined) {
                 syncedTotal++;
              }
            }
          } catch (err) {
            console.error(`Error processing consumer entry ${entry.name}`, err);
          }
        }));

        return new Response(JSON.stringify({ synced: syncedTotal }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' }
        });
      } catch (e: any) {
        console.error(e);
        return new Response('Error Processing Push', { status: 200 });
      }
    }

    const event = prEvent!;

    try {
      // Step 5: Set 'pending' commit status immediately
      await provider.setCommitStatus(
        eventOwner,
        eventRepo,
        event.headSha,
        'pending',
        'Substrate is checking...'
      );

      // Step 6: Fetch substrate.yaml from head commit
      let configContent = await provider.fetchFileContent(
        eventOwner,
        eventRepo,
        'substrate.yaml',
        event.headSha
      );

      let config: any = {};

      // Step 7: Zero-Config Heuristics if no substrate.yaml
      if (!configContent) {
        // First try to fetch org-level config
        const orgConfigContent = await provider.fetchFileContent(
          eventOwner,
          eventRepo,
          '.github/substrate-org.yaml',
          'main'
        );

        const prFiles = await provider.fetchPRFiles(eventOwner, eventRepo, event.prNumber);

        let matchedSchema = null;
        let matchedType = null;
        for (const file of prFiles) {
          if (file.includes('openapi')) {
            matchedSchema = file;
            matchedType = 'openapi';
            break;
          } else if (file.includes('schema.graphql')) {
            matchedSchema = file;
            matchedType = 'graphql';
            break;
          } else if (file.endsWith('.proto')) {
            matchedSchema = file;
            matchedType = 'protobuf';
            break;
          } else if (file.endsWith('schema.sql')) {
            matchedSchema = file;
            matchedType = 'sql';
            break;
          }
        }

        if (orgConfigContent || matchedSchema) {
          if (orgConfigContent) {
            configContent = orgConfigContent;
            config = parseYaml(orgConfigContent);

            // If global config is used, default spec to openapi.yaml
            if (!config.base_schema && !config.head_schema && !matchedSchema) {
              matchedSchema = 'openapi.yaml';
              matchedType = 'openapi';
            }
          } else {
            // Default config content from heuristics
            configContent = `service: auto-detected\nschema_type: ${matchedType}\nbase_schema: ${matchedSchema}\nhead_schema: ${matchedSchema}\n`;
            config = parseYaml(configContent);
          }

          // Merge heuristics
          if (!config.base_schema) config.base_schema = matchedSchema;
          if (!config.head_schema) config.head_schema = matchedSchema;
          if (!config.schema_type) config.schema_type = matchedType;
        } else {
          const missingConfigMsg = formatMissingConfigComment();
          await provider.postPRComment(eventOwner, eventRepo, event.prNumber, missingConfigMsg);

          // Ensure pending status remains (set neutrally above, spec says neutral/pending)
          await provider.setCommitStatus(
            eventOwner,
            eventRepo,
            event.headSha,
            'pending',
            'substrate.yaml not found — run `substrate init` to activate'
          );

          return new Response('Missing Config', { status: 200 });
        }
      } else {
        // Step 8: Parse config, fetch base + head spec files from GitHub API
        config = parseYaml(configContent);
      }
      let baseContent: string | null = null;
      let headContent: string | null = null;

      if (config.base_schema && config.head_schema) {
        baseContent = await provider.fetchFileContent(eventOwner, eventRepo, config.base_schema, event.baseBranch);
        headContent = await provider.fetchFileContent(eventOwner, eventRepo, config.head_schema, event.headSha);
      }

      if (!config.base_schema || !config.head_schema) {
        const configErrorMsg = "## ⚠️ Substrate — Config Error\n\nCould not parse `base_schema` and `head_schema` from your `substrate.yaml`.\n\n*Powered by [Substrate](https://github.com/KrushnaVardhanReddy/Substrate)*";
        await provider.postPRComment(eventOwner, eventRepo, event.prNumber, configErrorMsg);
        await provider.setCommitStatus(eventOwner, eventRepo, event.headSha, 'failure', 'Substrate config error — check substrate.yaml');
        return new Response('Config Error', { status: 200 });
      }

      if (baseContent === null && headContent !== null) {
        const welcomeMsg = "## 🎉 Welcome to Substrate!\n\nWe detected your new schema file. Since this is your first time adding it, there is no previous baseline to compare against.\n\nOnce this PR is merged, Substrate will begin monitoring all future pull requests for breaking changes!\n\n*Powered by [Substrate](https://github.com/KrushnaVardhanReddy/Substrate)*";
        await provider.postPRComment(eventOwner, eventRepo, event.prNumber, welcomeMsg);
        await provider.setCommitStatus(eventOwner, eventRepo, event.headSha, 'success', 'First-time setup detected — Ready to merge!');
        return new Response('First Time Setup', { status: 200 });
      }

      if (baseContent === null || headContent === null) {
        const configErrorMsg = "## ⚠️ Substrate — Fetch Error\n\nCould not fetch spec files from GitHub. Check that the paths match your repository structure.\n\n*Powered by [Substrate](https://github.com/KrushnaVardhanReddy/Substrate)*";
        await provider.postPRComment(eventOwner, eventRepo, event.prNumber, configErrorMsg);
        await provider.setCommitStatus(eventOwner, eventRepo, event.headSha, 'failure', 'Substrate fetch error — check file paths');
        return new Response('Config Error', { status: 200 });
      }

      // Step 9: REAL Container Service call
      let diffReport: DiffReport;
      try {
        const schemaType = config.schema_type || 'openapi';
        diffReport = await callContainerService(
          env.CONTAINER_SERVICE_URL,
          baseContent || '',
          headContent || '',
          configContent || '',
          schemaType
        );
      } catch (e: any) {
        console.error("Engine Call Failed:", e);
        await provider.setCommitStatus(eventOwner, eventRepo, event.headSha, 'failure', 'Substrate engine error — retry later');
        return new Response('Engine Error', { status: 200 });
      }

      // Save diff report to the API
      let previewToken: string | undefined;
      let diffId: string | undefined;
      const schemaType = config.schema_type || 'openapi';
      if (env.REGISTRY_API_URL) {
        try {
          const saveRes = await fetch(`${env.REGISTRY_API_URL}/api/v1/diff`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${env.REGISTRY_API_TOKEN}`
            },
            body: JSON.stringify({
              diff_report: diffReport,
              is_audit_mode: config.mode === 'audit',
              org: event.owner,
              provider_repo: event.fullName,
              pr_number: event.prNumber,
              commit_sha: event.headSha,
              head_schema_content: headContent,
              schema_type: schemaType,
              config_content: configContent,
              installation_id: event.installationId
            })
          });
          if (saveRes.ok) {
            const saveData = await saveRes.json() as any;
            diffId = saveData.id;
            previewToken = saveData.preview_token;
          } else {
            console.error(`Failed to save diff, status: ${saveRes.status}`);
          }
        } catch (e) {
          console.error("Failed to save diff:", e);
        }
      }

      // Step 9.5: Cross Repo Check
      let crossRepoResponse: CrossRepoCheckResponse | undefined;

      if (env.REGISTRY_API_URL) {
        const payload: CrossRepoCheckRequest = {
          installation_id: event.installationId,
          org: event.owner,
          provider_repo: event.fullName,
          head_schema_content: headContent || '',
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
      let commentBody = formatPRComment(diffReport, config, env.DASHBOARD_URL, event.owner, event.repo, event.prNumber, aiExplanation, aiSafePatch, previewToken);
      if (crossRepoSection) {
        commentBody += "\n" + crossRepoSection;
      }
      await provider.postPRComment(eventOwner, eventRepo, event.prNumber, commentBody);

      // Step 11: Set final commit status
      const statusState = getCommitStatusState(diffReport, config, crossRepoResponse);
      const statusDescription = getCommitStatusDescription(diffReport, crossRepoResponse, config);
      await provider.setCommitStatus(
        eventOwner,
        eventRepo,
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
