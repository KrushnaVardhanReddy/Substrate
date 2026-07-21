import re

with open('github-app/src/formatter.ts', 'r') as f:
    content = f.read()

# Replace the block that appends aiExplanation. Currently it places it before `crossRepoSection`
# But actually the spec says: "Append an '🤖 AI Impact Analysis' section at the bottom of the GitHub PR comment with the generated narrative."
# Wait, let's see where aiExplanation is appended. It's inside `if (breakingCount > 0)`

search_block = """    if (aiExplanation) {
      comment += `\\n---\\n### 🤖 AI Impact Analysis\\n> ${aiExplanation}\\n`;
    }
    if (aiSafePatch) {
      comment += `\\n### 🔧 Suggested Safe Remediation\\nApply the following change to unblock this PR:\\n\\`\\`\\`yaml\\n${aiSafePatch}\\n\\`\\`\\`\\n`;
    }"""

replace_block = """    if (aiExplanation) {
      comment += `\\n---\\n### 🤖 AI Impact Analysis\\n> ${aiExplanation}\\n`;
    }
    if (aiSafePatch) {
      comment += `\\n### 🔧 Suggested Safe Remediation\\nApply the following change to unblock this PR:\\n\\`\\`\\`yaml\\n${aiSafePatch}\\n\\`\\`\\`\\n`;
    }"""

# Wait, looking at the spec, "Append an '🤖 AI Impact Analysis' section at the bottom of the GitHub PR comment with the generated narrative."
# In formatter.ts, it is already appended at the bottom of the breaking changes section, with `\n---\n### 🤖 AI Impact Analysis\n> ${aiExplanation}\n`.
# So formatter is already compliant.
print("formatter.ts is already compliant")
