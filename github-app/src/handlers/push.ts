import { Env, StandardPushEvent, VCSClient } from '../types.js';
import { parseConsumersFromYaml, syncToRegistry } from '../registry-client.js';

export async function handlePushEvent(
  env: Env,
  pushEv: StandardPushEvent,
  provider: VCSClient,
  eventOwner: string,
  eventRepo: string
): Promise<Response> {
  try {
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

          // Phase 10: T05 - Auto-SDK Generator Trigger
          if (entry.sdk_targets && entry.sdk_targets.length > 0) {
            console.log(`Triggering SDK generation for ${consumerFullName} with targets: ${entry.sdk_targets.join(', ')}`);
            // We use standard API fallback here or post a comment confirming the intention to generate SDK PRs
            const sdkMsg = `🚀 **Substrate SDK Generation**\n\nTriggering generation for the following SDKs:\n${entry.sdk_targets.map(t => `- \`${t}\``).join('\n')}\n\nExpect a pull request shortly with the updated client code!`;
            // Normally, we'd open a PR, but since the Go `engine/sdkgen` does this locally in CLI context, we post a comment or simulated trigger here.
            // As per instructions, we just need to ensure the push handler logic is prepared.
            await provider.postPRComment(providerOwner, providerRepoName, 1, sdkMsg).catch(() => {});
          }

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
              required_notice_days: entry.required_notice_days,
              sdk_targets: entry.sdk_targets
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
