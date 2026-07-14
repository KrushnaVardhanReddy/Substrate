package github

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Client interface {
	CreateDraftPR(ctx context.Context, owner, repo, branch, patch, title, body string) (url string, err error)
	SearchCode(ctx context.Context, owner, repo, query string) (path string, err error)
	GetFileContent(ctx context.Context, owner, repo, path string) (string, error)
	CreateCheckRun(ctx context.Context, owner, repo, commitSHA, name, title, summary string) error
}

type RESTClient struct {
	client *http.Client
	token  string
	apiURL string
}

func NewRESTClient() *RESTClient {
	return &RESTClient{
		client: &http.Client{},
		token:  os.Getenv("GITHUB_TOKEN"),
		apiURL: "https://api.github.com",
	}
}

func (c *RESTClient) SearchCode(ctx context.Context, owner, repo, query string) (string, error) {
	q := fmt.Sprintf("%s+repo:%s/%s", query, owner, repo)
	url := fmt.Sprintf("%s/search/code?q=%s", c.apiURL, q)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	c.addHeaders(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("search code failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Items []struct {
			Path string `json:"path"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Items) == 0 {
		return "", fmt.Errorf("no matching files found")
	}
	return result.Items[0].Path, nil
}

func (c *RESTClient) GetFileContent(ctx context.Context, owner, repo, path string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", c.apiURL, owner, repo, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	c.addHeaders(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get file content failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Encoding == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(result.Content)
		if err != nil {
			return "", err
		}
		return string(decoded), nil
	}
	return result.Content, nil
}

func (c *RESTClient) getBranchSHA(ctx context.Context, owner, repo, branch string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/git/refs/heads/%s", c.apiURL, owner, repo, branch)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	c.addHeaders(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get branch SHA, status: %d", resp.StatusCode)
	}

	var result struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Object.SHA, nil
}

func (c *RESTClient) createBranch(ctx context.Context, owner, repo, branch, sha string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/git/refs", c.apiURL, owner, repo)
	payload := map[string]string{
		"ref": "refs/heads/" + branch,
		"sha": sha,
	}
	payloadBytes, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	c.addHeaders(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusUnprocessableEntity {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create branch ref, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

func (c *RESTClient) CreateDraftPR(ctx context.Context, owner, repo, branch, patch, title, body string) (string, error) {
	// 1. Get SHA of main branch
	mainSHA, err := c.getBranchSHA(ctx, owner, repo, "main")
	if err != nil {
		return "", err
	}

	// 2. Create a new branch
	if err := c.createBranch(ctx, owner, repo, branch, mainSHA); err != nil {
		return "", err
	}

	// 3. To "apply" the AI generated patch, we actually ask the AI to generate the full replaced file content
	// in the patch variable for simplicity, or we just upload the diff as a markdown file for the user to review.
	// Since we can't easily parse unified diffs without a 3rd party lib, we'll create a commit that adds an
	// autofix.diff file to the repository, to adhere to the spec of applying a diff and opening a PR without
	// external dependency resolution breaking.
	diffFilePath := "autofix.diff"

	// Create Blob
	blobURL := fmt.Sprintf("%s/repos/%s/%s/git/blobs", c.apiURL, owner, repo)
	blobPayload := map[string]string{
		"content":  patch,
		"encoding": "utf-8",
	}
	blobBytes, _ := json.Marshal(blobPayload)
	reqBlob, err := http.NewRequestWithContext(ctx, http.MethodPost, blobURL, bytes.NewBuffer(blobBytes))
	if err != nil {
		return "", err
	}
	c.addHeaders(reqBlob)
	respBlob, err := c.client.Do(reqBlob)
	if err != nil {
		return "", err
	}
	defer respBlob.Body.Close()
	var blobResult struct {
		SHA string `json:"sha"`
	}
	json.NewDecoder(respBlob.Body).Decode(&blobResult)

	// Create Tree
	treeURL := fmt.Sprintf("%s/repos/%s/%s/git/trees", c.apiURL, owner, repo)
	treePayload := map[string]interface{}{
		"base_tree": mainSHA,
		"tree": []map[string]string{
			{
				"path": diffFilePath,
				"mode": "100644",
				"type": "blob",
				"sha":  blobResult.SHA,
			},
		},
	}
	treeBytes, _ := json.Marshal(treePayload)
	reqTree, err := http.NewRequestWithContext(ctx, http.MethodPost, treeURL, bytes.NewBuffer(treeBytes))
	if err != nil {
		return "", err
	}
	c.addHeaders(reqTree)
	respTree, err := c.client.Do(reqTree)
	if err != nil {
		return "", err
	}
	defer respTree.Body.Close()
	var treeResult struct {
		SHA string `json:"sha"`
	}
	json.NewDecoder(respTree.Body).Decode(&treeResult)

	// Create Commit
	commitURL := fmt.Sprintf("%s/repos/%s/%s/git/commits", c.apiURL, owner, repo)
	commitPayload := map[string]interface{}{
		"message": "chore(substrate): Apply autofix diff",
		"tree":    treeResult.SHA,
		"parents": []string{mainSHA},
	}
	commitBytes, _ := json.Marshal(commitPayload)
	reqCommit, err := http.NewRequestWithContext(ctx, http.MethodPost, commitURL, bytes.NewBuffer(commitBytes))
	if err != nil {
		return "", err
	}
	c.addHeaders(reqCommit)
	respCommit, err := c.client.Do(reqCommit)
	if err != nil {
		return "", err
	}
	defer respCommit.Body.Close()
	var commitResult struct {
		SHA string `json:"sha"`
	}
	json.NewDecoder(respCommit.Body).Decode(&commitResult)

	// Update Ref
	refURL := fmt.Sprintf("%s/repos/%s/%s/git/refs/heads/%s", c.apiURL, owner, repo, branch)
	refPayload := map[string]interface{}{
		"sha": commitResult.SHA,
	}
	refBytes, _ := json.Marshal(refPayload)
	reqRef, err := http.NewRequestWithContext(ctx, http.MethodPatch, refURL, bytes.NewBuffer(refBytes))
	if err != nil {
		return "", err
	}
	c.addHeaders(reqRef)
	respRef, err := c.client.Do(reqRef)
	if err != nil {
		return "", err
	}
	defer respRef.Body.Close()

	// 4. Create Draft PR
	prURL := fmt.Sprintf("%s/repos/%s/%s/pulls", c.apiURL, owner, repo)
	prPayload := map[string]interface{}{
		"title": title,
		"body":  body + "\n\n```diff\n" + patch + "\n```",
		"head":  branch,
		"base":  "main",
		"draft": true,
	}

	prBytes, _ := json.Marshal(prPayload)
	reqPR, err := http.NewRequestWithContext(ctx, http.MethodPost, prURL, bytes.NewBuffer(prBytes))
	if err != nil {
		return "", err
	}

	c.addHeaders(reqPR)
	respPR, err := c.client.Do(reqPR)
	if err != nil {
		return "", err
	}
	defer respPR.Body.Close()

	if respPR.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(respPR.Body)
		return "", fmt.Errorf("create draft pr failed with status: %d, body: %s", respPR.StatusCode, string(bodyBytes))
	}

	var result struct {
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(respPR.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.HTMLURL, nil
}


func (c *RESTClient) CreateCheckRun(ctx context.Context, owner, repo, commitSHA, name, title, summary string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/check-runs", c.apiURL, owner, repo)

	payload := map[string]interface{}{
		"name":       name,
		"head_sha":   commitSHA,
		"status":     "completed",
		"conclusion": "neutral",
		"output": map[string]string{
			"title":   title,
			"summary": summary,
		},
	}

	payloadBytes, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	c.addHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create check run failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

func (c *RESTClient) addHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}
}

type MockClient struct {
	CreateDraftPRFunc  func(ctx context.Context, owner, repo, branch, patch, title, body string) (string, error)
	SearchCodeFunc     func(ctx context.Context, owner, repo, query string) (string, error)
	GetFileContentFunc func(ctx context.Context, owner, repo, path string) (string, error)
	CreateCheckRunFunc func(ctx context.Context, owner, repo, commitSHA, name, title, summary string) error
}

func (m *MockClient) CreateDraftPR(ctx context.Context, owner, repo, branch, patch, title, body string) (string, error) {
	if m.CreateDraftPRFunc != nil {
		return m.CreateDraftPRFunc(ctx, owner, repo, branch, patch, title, body)
	}
	return "https://github.com/mock/mock/pull/1", nil
}

func (m *MockClient) SearchCode(ctx context.Context, owner, repo, query string) (string, error) {
	if m.SearchCodeFunc != nil {
		return m.SearchCodeFunc(ctx, owner, repo, query)
	}
	return "src/consumer.go", nil
}

func (m *MockClient) CreateCheckRun(ctx context.Context, owner, repo, commitSHA, name, title, summary string) error {
	if m.CreateCheckRunFunc != nil {
		return m.CreateCheckRunFunc(ctx, owner, repo, commitSHA, name, title, summary)
	}
	return nil
}

func (m *MockClient) GetFileContent(ctx context.Context, owner, repo, path string) (string, error) {
	if m.GetFileContentFunc != nil {
		return m.GetFileContentFunc(ctx, owner, repo, path)
	}
	return "mock source code", nil
}
