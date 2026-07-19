package github

import (
	"fmt"
	"strings"
)

func GeneratePRComment(chaosOutput string) string {
	return fmt.Sprintf("## Proof of Breakage\n\n```text\n%s\n```\n", chaosOutput)
}

func GenerateNegotiationComment(chaosOutput string, mentions []string) string {
	mentionsStr := strings.Join(mentions, ", ")
	return fmt.Sprintf("## Action Required: Breaking Change Detected\n\nSubstrate has detected a breaking change in this PR. Affected consumer leads have been tagged for review.\n\n%s\n\nPlease review the changes. React with 👍 to acknowledge and approve this breaking change, or 👎 to block.\n\n## Proof of Breakage\n\n```text\n%s\n```\n", mentionsStr, chaosOutput)
}
func ParseNegotiationComment(comment string) []string {
	// A simple helper if we need it later to parse out mentions from the comment body
	var mentions []string
	lines := strings.Split(comment, "\n")
	for _, line := range lines {
		if strings.Contains(line, "@") {
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "@") {
					mentions = append(mentions, part)
				}
			}
		}
	}
	return mentions
}
