package mockgen

import (
	"context"
	"errors"
	"github.com/spf13/viper"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
	"github.com/KrushnaVardhanReddy/substrate/engine/pkg/ai"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

func TestDiscoverFixtures(t *testing.T) {
	tempDir := t.TempDir()

	files := []string{
		"__fixtures__/test1.json",
		"__fixtures__/sub/test2.json",
		"testdata/test3.json",
		"testdata/sub/test4.json",
		"other/test5.json",
		"__fixtures__/notjson.txt",
	}

	for _, file := range files {
		fullPath := filepath.Join(tempDir, file)
		assert.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
		assert.NoError(t, os.WriteFile(fullPath, []byte("{}"), 0644))
	}

	fixtures, err := DiscoverFixtures(tempDir)
	assert.NoError(t, err)
	assert.Len(t, fixtures, 4)

	var foundPaths []string
	for _, f := range fixtures {
		foundPaths = append(foundPaths, filepath.ToSlash(f))
	}
	joined := strings.Join(foundPaths, " ")
	assert.Contains(t, joined, "__fixtures__/test1.json")
	assert.Contains(t, joined, "__fixtures__/sub/test2.json")
	assert.Contains(t, joined, "testdata/test3.json")
	assert.Contains(t, joined, "testdata/sub/test4.json")
	assert.NotContains(t, joined, "other/test5.json")
	assert.NotContains(t, joined, "__fixtures__/notjson.txt")
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "pure json",
			input:    `{"a": 1}`,
			expected: `{"a": 1}`,
		},
		{
			name:     "markdown json block",
			input:    "```json\n{\"a\": 1}\n```",
			expected: `{"a": 1}`,
		},
		{
			name:     "markdown block without language",
			input:    "```\n{\"b\": 2}\n```",
			expected: `{"b": 2}`,
		},
		{
			name:     "with surrounding whitespace",
			input:    "\n```json\n  {\"c\": 3}  \n```\n",
			expected: `{"c": 3}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSON(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

type mockAIStream struct {
	chunks []string
	idx    int
}

func (m *mockAIStream) Recv() (string, error) {
	if m.idx >= len(m.chunks) {
		return "", io.EOF
	}
	res := m.chunks[m.idx]
	m.idx++
	return res, nil
}

func (m *mockAIStream) Close() error {
	return nil
}

type mockAIClient struct {
	response string
	err      error
	lastReq  openai.ChatCompletionRequest
}

func (m *mockAIClient) CreateChatCompletionStream(ctx context.Context, req openai.ChatCompletionRequest) (ai.AIChatCompletionStream, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.lastReq = req
	return &mockAIStream{chunks: []string{m.response}}, nil
}

func TestUpdateFixture(t *testing.T) {
	ctx := context.Background()
	diffReport := &report.DiffReport{
		Summary: report.Summary{TotalChanges: 1},
	}
	fixtureContent := []byte(`{"old": "data"}`)

	mockResponse := "```json\n{\"new\": \"data\"}\n```"
	client := &mockAIClient{response: mockResponse}

	updated, err := UpdateFixture(ctx, client, diffReport, fixtureContent)
	assert.NoError(t, err)
	assert.Equal(t, `{"new": "data"}`, string(updated))

	// Verify the prompt contained the diff and the fixture
	assert.Len(t, client.lastReq.Messages, 2)
	assert.Equal(t, openai.ChatMessageRoleSystem, client.lastReq.Messages[0].Role)
	assert.Contains(t, client.lastReq.Messages[0].Content, "You are a QA mock data generator")

	userPrompt := client.lastReq.Messages[1].Content
	assert.Contains(t, userPrompt, "DiffReport:")
	assert.Contains(t, userPrompt, "Current Fixture JSON:\n{\"old\": \"data\"}")

	// Test AI client error
	clientErr := &mockAIClient{err: errors.New("ai error")}
	_, err = UpdateFixture(ctx, clientErr, diffReport, fixtureContent)
	assert.ErrorContains(t, err, "ai error")
}

func TestRunMockGenerator_Errors(t *testing.T) {
	ctx := context.Background()
	// Pass non-existent paths to trigger compare error
	err := RunMockGenerator(ctx, "/bad/repo", "/bad/old", "/bad/new", "owner", "repo", "token", "branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to compare schema")
}

func TestCreatePR_NoToken(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	// Create PR without git repo should fail to open repo
	err := CreatePR(ctx, tempDir, "branch", "owner", "repo", "token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open repo")
}

func TestCreatePR_CleanRepo(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	// Initialize a real git repo
	r, err := git.PlainInit(tempDir, false)
	assert.NoError(t, err)

	w, err := r.Worktree()
	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("hello"), 0644)
	assert.NoError(t, err)

	_, err = w.Add("test.txt")
	assert.NoError(t, err)

	_, err = w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err)

	// Since there are no new changes, CreatePR should return nil immediately (clean repo)
	err = CreatePR(ctx, tempDir, "branch", "owner", "repo", "token")
	assert.NoError(t, err)
}

func TestCreatePR_WithChanges_NoToken(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	// Initialize a real git repo
	r, err := git.PlainInit(tempDir, false)
	assert.NoError(t, err)

	w, err := r.Worktree()
	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("hello"), 0644)
	assert.NoError(t, err)

	_, err = w.Add("test.txt")
	assert.NoError(t, err)

	_, err = w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err)

	// Make a change
	err = os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("hello world"), 0644)
	assert.NoError(t, err)

	// CreatePR with empty token should commit but skip PR creation
	err = CreatePR(ctx, tempDir, "branch", "owner", "repo", "")
	assert.NoError(t, err)

	// Verify it committed
	status, err := w.Status()
	assert.NoError(t, err)
	assert.True(t, status.IsClean())
}

func TestCreatePR_API(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	// Initialize a real git repo
	r, err := git.PlainInit(tempDir, false)
	assert.NoError(t, err)

	w, err := r.Worktree()
	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("hello"), 0644)
	assert.NoError(t, err)

	_, err = w.Add("test.txt")
	assert.NoError(t, err)

	_, err = w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err)

	// Make a change
	err = os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("hello world"), 0644)
	assert.NoError(t, err)

	// Try a bad token call which should fail from real API or timeout/dns
	err = CreatePR(ctx, tempDir, "branch", "KrushnaVardhanReddy", "substrate", "bad-token")
	// If it tries to hit actual github it will get a 401 or fail in CI environment.
	if err != nil {
		assert.Contains(t, err.Error(), "github API error: status 401")
	}
}

func TestRunMockGenerator_Success(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	// Need old/new schemas
	oldSchema := filepath.Join(tempDir, "old.yaml")
	newSchema := filepath.Join(tempDir, "new.yaml")

	schemaData := `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths: {}`

	os.WriteFile(oldSchema, []byte(schemaData), 0644)
	os.WriteFile(newSchema, []byte(schemaData), 0644)

	// Since we can't easily mock the ai.NewAIClient inside RunMockGenerator (it instantiates directly),
	// it will likely fail with missing API key or unsupported provider, depending on env vars.
	// But we can test up to the AI client error.
	os.Setenv("SUBSTRATE_AI_PROVIDER", "unsupported-test")
	err := RunMockGenerator(ctx, tempDir, oldSchema, newSchema, "owner", "repo", "token", "branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create AI client")
}

func TestUpdateFixture_JSONExtractFallback(t *testing.T) {
	ctx := context.Background()
	diffReport := &report.DiffReport{}

	// No codeblocks response
	mockResponse := `{"new": "data"}`
	client := &mockAIClient{response: mockResponse}

	updated, err := UpdateFixture(ctx, client, diffReport, []byte("{}"))
	assert.NoError(t, err)
	assert.Equal(t, `{"new": "data"}`, string(updated))
}

func TestCreatePR_StatusError(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	err := CreatePR(ctx, tempDir, "branch", "owner", "repo", "token")
	assert.Error(t, err)
	// it will fail to open repo
}

func TestRunMockGenerator_BadFixtures(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()

	oldSchema := filepath.Join(tempDir, "old.yaml")
	newSchema := filepath.Join(tempDir, "new.yaml")

	schemaData := `openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths: {}`

	os.WriteFile(oldSchema, []byte(schemaData), 0644)
	os.WriteFile(newSchema, []byte(schemaData), 0644)

	err := RunMockGenerator(ctx, "/does-not-exist-12345", oldSchema, newSchema, "owner", "repo", "token", "branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to discover fixtures")
}

func init() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}
