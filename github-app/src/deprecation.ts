// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { createIssue, findOpenIssue, closeIssue } from './github-client';
import { Deprecation, CrossRepoCheckResponse, ConsumerResult } from './types';

export async function processDeprecations(
  token: string,
  deprecations: Deprecation[],
  crossRepoResponse: CrossRepoCheckResponse | undefined
) {
  if (!crossRepoResponse || !crossRepoResponse.results) {
    return;
  }

  const consumers = crossRepoResponse.results;

  for (const dep of deprecations) {
    const titlePrefix = `Deprecation Notice: ${dep.endpoint}`;

    for (const consumer of consumers) {
      const parts = consumer.consumer_repo.split('/');
      if (parts.length !== 2) continue;
      const [owner, repo] = parts;

      // Check if this consumer actually uses the deprecated endpoint
      // Usage would show up in warnings or breaking changes with a rule like ENDPOINT_DEPRECATED
      let isUsing = false;
      const allChanges = [
        ...(consumer.diff_report.warnings || []),
        ...(consumer.diff_report.breaking_changes || []),
      ];

      for (const change of allChanges) {
        if (
          (change.rule_id === 'ENDPOINT_DEPRECATED' || change.rule_id === 'FIELD_DEPRECATED') &&
          change.path.includes(dep.endpoint)
        ) {
          isUsing = true;
          break;
        }
      }

      try {
        const searchRes = await findOpenIssue(token, owner, repo, titlePrefix);
        const issues = searchRes.items || [];
        const hasOpenIssue = issues.length > 0;

        if (isUsing) {
          // If usage > 0, we open the issue if it doesn't exist
          if (!hasOpenIssue) {
            const body = `The endpoint \`${dep.endpoint}\` is scheduled for deprecation on **${dep.sunset_date}**.\n\nPlease migrate to an alternative endpoint before this date to avoid disruption.\n\n_This issue was opened automatically by Substrate._`;
            await createIssue(token, owner, repo, titlePrefix, body);
            console.log(`Successfully opened deprecation issue in ${consumer.consumer_repo}`);
          } else {
            console.log(`Deprecation issue already open in ${consumer.consumer_repo}`);
          }
        } else {
          // If usage == 0, we close the issue if it exists
          if (hasOpenIssue) {
            for (const issue of issues) {
              const comment = `Usage of \`${dep.endpoint}\` has successfully dropped to 0%. Great job! Closing this deprecation notice.`;
              await closeIssue(token, owner, repo, issue.number, comment);
              console.log(`Successfully closed deprecation issue in ${consumer.consumer_repo}`);
            }
          }
        }
      } catch (error) {
        console.error(`Failed to process deprecation for ${consumer.consumer_repo}:`, error);
      }
    }
  }
}
