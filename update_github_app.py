import re

with open('github-app/src/index.ts', 'r') as f:
    content = f.read()

# Add AIImpact API call logic
impact_code = """      // P9-T09 AI Impact Analysis
      if (diffReport.summary.breaking_count > 0 && env.REGISTRY_API_URL) {
        try {
          const impactReq = {
            changes: diffReport.breaking_changes.map(c => ({ rule_id: c.rule_id, path: c.path, description: c.description })),
            consumers: crossRepoResponse && crossRepoResponse.results ? crossRepoResponse.results.map(r => ({ name: r.consumer_repo })) : []
          };

          const impactRes = await fetch(`${env.REGISTRY_API_URL}/api/v1/ai/impact`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${env.REGISTRY_API_TOKEN}`
            },
            body: JSON.stringify(impactReq)
          });

          if (impactRes.ok) {
            const impactData = await impactRes.json() as any;
            if (impactData && impactData.explanation) {
              aiExplanation = impactData.explanation;
            }
          }
        } catch (e) {
          console.error("AI Impact request failed:", e);
        }
      }
"""

content = re.sub(
    r'(// Step 9\.5: Fetch Risk Score)',
    impact_code + r'\n      \1',
    content
)

with open('github-app/src/index.ts', 'w') as f:
    f.write(content)

print("Updated github-app/src/index.ts")
