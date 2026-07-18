package discovery

import (
	"context"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/github"
)

type RepositoryScanner struct {
	ghClient github.Client
}

func NewRepositoryScanner(ghClient github.Client) *RepositoryScanner {
	return &RepositoryScanner{
		ghClient: ghClient,
	}
}

var targetFiles = []string{
	"package.json",
	"go.mod",
	"docker-compose.yml",
	"docker-compose.yaml",
	"openapi-generator-config.yaml",
	".openapi-generator-config.yaml",
	"openapitools.json",
}

func (s *RepositoryScanner) ScanRepository(ctx context.Context, owner, repo string, urlResolver func(string) string) ([]DependencyEdge, error) {
	files := make(map[string][]byte)

	for _, filename := range targetFiles {
		content, err := s.ghClient.GetFileContent(ctx, owner, repo, filename)
		if err == nil && content != "" {
			files[filename] = []byte(content)
		}
	}

	aggregator := NewAggregator()
	repoFullName := owner + "/" + repo

	err := aggregator.ScanAndAggregate(ctx, repoFullName, files, urlResolver)
	if err != nil {
		return nil, err
	}

	return aggregator.GetEdges(), nil
}
