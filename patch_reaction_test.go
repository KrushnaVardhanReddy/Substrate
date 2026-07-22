package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	content, err := os.ReadFile("api/internal/webhook/reaction_test.go")
	if err != nil {
		panic(err)
	}

	str := string(content)
	str = strings.Replace(str, `	checkRunCalled := false
	mockGh.CreateCheckRunFunc = func(ctx context.Context, owner, repo, commitSHA, name, title, summary string) error {
		checkRunCalled = true
		return nil
	}`, `	checkRunCalled := false
	mockGh.CreateCheckRunFunc = func(ctx context.Context, owner, repo, commitSHA, name, title, summary string) error {
		checkRunCalled = true
		return nil
	}
	mockGh.GetIssueCommentReactionsFunc = func(ctx context.Context, owner, repo string, issueNumber int, commentID int64) ([]string, error) {
		return []string{"@alice"}, nil
	}`, 1)

	str = strings.Replace(str, `"comment": {"body": "## Action Required\n@alice"}`, `"comment": {"id": 123, "body": "## Action Required\n@alice"}`, 1)

	err = os.WriteFile("api/internal/webhook/reaction_test.go", []byte(str), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("patched reaction_test.go")
}
