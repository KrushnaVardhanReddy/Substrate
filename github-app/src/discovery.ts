// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { Env, InstallationRepo } from './types.js';
import {
  generateInstallationToken,
  checkFileExists,
  getDefaultBranch,
  getBranchSha,
  createBranch,
  createCommitWithFile,
  createPullRequest
} from './github-client.js';

const SPEC_FILES = ['openapi.yaml', 'schema.graphql', '.env'];
const INIT_BRANCH = 'substrate-init';

function generateSubstrateYaml(specFile: string): string {
  let schemaType = 'openapi';
  if (specFile === 'schema.graphql') schemaType = 'graphql';
  else if (specFile === '.env') schemaType = 'dotenv';

  return `schema_type: ${schemaType}
base_schema: ${specFile}
head_schema: ${specFile}
on_breaking_change: block
`;
}

export async function processAutoDiscovery(
  env: Env,
  installationId: number,
  repositories: InstallationRepo[]
): Promise<void> {
  let token: string;
  try {
    token = await generateInstallationToken(
      env.GITHUB_APP_ID,
      env.GITHUB_APP_PRIVATE_KEY,
      installationId
    );
  } catch (e) {
    console.error(`Auto-Discovery: Failed to generate token for installation ${installationId}`, e);
    return;
  }

  for (const repo of repositories) {
    const owner = repo.owner.login || repo.full_name.split('/')[0];
    const repoName = repo.name;

    try {
      const defaultBranch = await getDefaultBranch(token, owner, repoName);

      let foundSpecFile: string | null = null;
      for (const file of SPEC_FILES) {
        const exists = await checkFileExists(token, owner, repoName, file, defaultBranch);
        if (exists) {
          foundSpecFile = file;
          break;
        }
      }

      if (foundSpecFile) {
        console.log(`Auto-Discovery: Found ${foundSpecFile} in ${owner}/${repoName}`);

        // Check if substrate.yaml already exists on default branch
        const substrateExists = await checkFileExists(token, owner, repoName, 'substrate.yaml', defaultBranch);
        if (substrateExists) {
            console.log(`Auto-Discovery: substrate.yaml already exists in ${owner}/${repoName}, skipping.`);
            continue;
        }

        const baseSha = await getBranchSha(token, owner, repoName, defaultBranch);

        // create branch
        await createBranch(token, owner, repoName, INIT_BRANCH, baseSha);

        const yamlContent = generateSubstrateYaml(foundSpecFile);

        // create commit
        await createCommitWithFile(
          token,
          owner,
          repoName,
          INIT_BRANCH,
          'substrate.yaml',
          yamlContent,
          'Initialize Substrate configuration'
        );

        // open PR
        const prBody = `## 🚀 Substrate Zero-Touch Onboarding\n\nWe detected a \`${foundSpecFile}\` file in your repository! This PR injects a default \`substrate.yaml\` configuration so you can start catching schema breaking changes instantly.\n\nMerge this PR to activate Substrate for this repository.`;

        const prNumber = await createPullRequest(
          token,
          owner,
          repoName,
          'Initialize Substrate Configuration',
          INIT_BRANCH,
          defaultBranch,
          prBody
        );

        if (prNumber) {
           console.log(`Auto-Discovery: Created PR #${prNumber} for ${owner}/${repoName}`);
        } else {
           console.log(`Auto-Discovery: PR already exists or failed for ${owner}/${repoName}`);
        }
      }

    } catch (e) {
       console.error(`Auto-Discovery: Error processing repo ${owner}/${repoName}`, e);
    }
  }
}
