package github

import "fmt"

func GeneratePRComment(chaosOutput string) string {
	return fmt.Sprintf("## Proof of Breakage\n\n```text\n%s\n```\n", chaosOutput)
}
