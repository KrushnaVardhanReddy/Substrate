package helpers

import (
	"context"
	"log"
	"strings"

	"github.com/google/go-github/v62/github"
)

func SeedFile(ctx context.Context, client *github.Client, owner, repo, path, branch, content string) {
	fileContent, _, _, err := client.Repositories.GetContents(ctx, owner, repo, path, &github.RepositoryContentGetOptions{Ref: branch})
	
	var sha *string
	if err == nil {
		sha = fileContent.SHA
	}

	opts := &github.RepositoryContentFileOptions{
		Message: github.String("chore(e2e): seed baseline " + path),
		Content: []byte(content),
		Branch:  github.String(branch),
		SHA:     sha,
	}

	_, _, err = client.Repositories.UpdateFile(ctx, owner, repo, path, opts)
	if err != nil && !strings.Contains(err.Error(), "does not match the current") {
		// Ignore if it's already the exact same content, else log
		log.Printf("Note: %v", err)
	}
}
