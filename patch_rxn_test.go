//go:build ignore

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
	str = strings.Replace(str, `	mockGh.CreateCheckRunFunc = func(ctx context.Context, owner, repo, commitSHA, name, title, summary string) error {
		checkRunCalled = true
		return nil
	}`, `	mockGh.CreateCheckRunFunc = func(ctx context.Context, owner, repo, commitSHA, name, title, summary string) error {
		checkRunCalled = true
		return nil
	}
	mockGh.GetPullRequestHeadSHAFunc = func(ctx context.Context, owner, repo string, issueNumber int) (string, error) {
		return "mock_head_sha", nil
	}`, 1)

	err = os.WriteFile("api/internal/webhook/reaction_test.go", []byte(str), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("patched reaction_test.go")
}
