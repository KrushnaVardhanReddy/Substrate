// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package github

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type CheckRun struct {
	Name       string `json:"name"`
	Conclusion string `json:"conclusion"`
}

type Client interface {
	RequestReviewers(ctx context.Context, owner, repo string, pullNumber int, reviewers []string) error
	CreateDraftPR(ctx context.Context, owner, repo, branch, patch, title, body string) (url string, err error)
	CreatePR(ctx context.Context, owner, repo, title, branch, body string) (url string, err error)
	CommitAndPushFile(ctx context.Context, owner, repo, branch, filePath, fileContent, commitMessage string) error
	SearchCode(ctx context.Context, owner, repo, query string) (path string, err error)
	GetFileContent(ctx context.Context, owner, repo, path string) (string, error)
	CreateCheckRun(ctx context.Context, owner, repo, commitSHA, name, title, summary, conclusion string) error
	CreatePendingCheckRun(ctx context.Context, owner, repo, commitSHA, name, title, summary string) error
	GetIssueCommentReactions(ctx context.Context, owner, repo string, issueNumber int, commentID int64) ([]string, error)
	CreateIssueComment(ctx context.Context, owner, repo string, issueNumber int, body string) error
	ListIssueComments(ctx context.Context, owner, repo string, issueNumber int) ([]string, error)
	GetPullRequestHeadSHA(ctx context.Context, owner, repo string, issueNumber int) (string, error)
	ListCheckRunsForRef(ctx context.Context, owner, repo, ref string) ([]CheckRun, error)
}

type RESTClient struct {
	client *http.Client
	token  string
	apiURL string
}

func NewRESTClient() *RESTClient {
	apiURL := os.Getenv("GITHUB_API_URL")
	if apiURL == "" {
		apiURL = "https://api.github.com"
	}
	return &RESTClient{
		client: &http.Client{},
		token:  os.Getenv("GITHUB_TOKEN"),
		apiURL: apiURL,
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
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("GetFileContent %s returned %d: %s", url, resp.StatusCode, string(bodyBytes))
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

func (c *RESTClient) GetPullRequestHeadSHA(ctx context.Context, owner, repo string, issueNumber int) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", c.apiURL, owner, repo, issueNumber)
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
		return "", fmt.Errorf("get pull request failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Head struct {
			SHA string `json:"sha"`
		} `json:"head"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Head.SHA, nil
}

func (c *RESTClient) CreateCheckRun(ctx context.Context, owner, repo, commitSHA, name, title, summary, conclusion string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/check-runs", c.apiURL, owner, repo)

	payload := map[string]interface{}{
		"name":       name,
		"head_sha":   commitSHA,
		"status":     "completed",
		"conclusion": conclusion,
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

func (c *RESTClient) CreatePendingCheckRun(ctx context.Context, owner, repo, commitSHA, name, title, summary string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/check-runs", c.apiURL, owner, repo)

	payload := map[string]interface{}{
		"name":     name,
		"head_sha": commitSHA,
		"status":   "in_progress",
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

func (c *RESTClient) GetIssueCommentReactions(ctx context.Context, owner, repo string, issueNumber int, commentID int64) ([]string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/comments/%d/reactions", c.apiURL, owner, repo, commentID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.addHeaders(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get issue comment reactions failed with status: %d", resp.StatusCode)
	}

	var reactions []struct {
		Content string `json:"content"`
		User    struct {
			Login string `json:"login"`
		} `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&reactions); err != nil {
		return nil, err
	}

	var users []string
	for _, r := range reactions {
		if r.Content == "+1" {
			users = append(users, "@"+r.User.Login)
		}
	}
	return users, nil
}

func (c *RESTClient) addHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}
}

type MockClient struct {
	RequestReviewersFunc         func(ctx context.Context, owner, repo string, pullNumber int, reviewers []string) error
	CreateDraftPRFunc            func(ctx context.Context, owner, repo, branch, patch, title, body string) (string, error)
	SearchCodeFunc               func(ctx context.Context, owner, repo, query string) (string, error)
	GetFileContentFunc           func(ctx context.Context, owner, repo, path string) (string, error)
	CreateCheckRunFunc           func(ctx context.Context, owner, repo, commitSHA, name, title, summary, conclusion string) error
	CreatePendingCheckRunFunc    func(ctx context.Context, owner, repo, commitSHA, name, title, summary string) error
	GetIssueCommentReactionsFunc func(ctx context.Context, owner, repo string, issueNumber int, commentID int64) ([]string, error)
	GetPullRequestHeadSHAFunc    func(ctx context.Context, owner, repo string, issueNumber int) (string, error)
	ListCheckRunsForRefFunc      func(ctx context.Context, owner, repo, ref string) ([]CheckRun, error)
}

func (m *MockClient) RequestReviewers(ctx context.Context, owner, repo string, pullNumber int, reviewers []string) error {
	if m.RequestReviewersFunc != nil {
		return m.RequestReviewersFunc(ctx, owner, repo, pullNumber, reviewers)
	}
	return nil
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

func (m *MockClient) CreatePendingCheckRun(ctx context.Context, owner, repo, commitSHA, name, title, summary string) error {
	if m.CreatePendingCheckRunFunc != nil {
		return m.CreatePendingCheckRunFunc(ctx, owner, repo, commitSHA, name, title, summary)
	}
	return nil
}

func (c *RESTClient) CreateIssueComment(ctx context.Context, owner, repo string, issueNumber int, body string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", c.apiURL, owner, repo, issueNumber)
	payload := map[string]string{"body": body}
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
		return fmt.Errorf("failed to create issue comment, status: %d", resp.StatusCode)
	}
	return nil
}

func (c *RESTClient) ListIssueComments(ctx context.Context, owner, repo string, issueNumber int) ([]string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", c.apiURL, owner, repo, issueNumber)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.addHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list issue comments, status: %d", resp.StatusCode)
	}

	var comments []struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		return nil, err
	}

	var bodies []string
	for _, c := range comments {
		bodies = append(bodies, c.Body)
	}
	return bodies, nil
}

func (m *MockClient) CreateIssueComment(ctx context.Context, owner, repo string, issueNumber int, body string) error {
	return nil
}

func (m *MockClient) ListIssueComments(ctx context.Context, owner, repo string, issueNumber int) ([]string, error) {
	return []string{}, nil
}

func (m *MockClient) GetIssueCommentReactions(ctx context.Context, owner, repo string, issueNumber int, commentID int64) ([]string, error) {
	if m.GetIssueCommentReactionsFunc != nil {
		return m.GetIssueCommentReactionsFunc(ctx, owner, repo, issueNumber, commentID)
	}
	return []string{}, nil
}

func (m *MockClient) GetPullRequestHeadSHA(ctx context.Context, owner, repo string, issueNumber int) (string, error) {
	if m.GetPullRequestHeadSHAFunc != nil {
		return m.GetPullRequestHeadSHAFunc(ctx, owner, repo, issueNumber)
	}
	return "mock_head_sha", nil
}

func (m *MockClient) CreateCheckRun(ctx context.Context, owner, repo, commitSHA, name, title, summary, conclusion string) error {
	if m.CreateCheckRunFunc != nil {
		return m.CreateCheckRunFunc(ctx, owner, repo, commitSHA, name, title, summary, conclusion)
	}
	return nil
}

func (m *MockClient) GetFileContent(ctx context.Context, owner, repo, path string) (string, error) {
	if m.GetFileContentFunc != nil {
		return m.GetFileContentFunc(ctx, owner, repo, path)
	}
	return "mock source code", nil
}

func (c *RESTClient) ListCheckRunsForRef(ctx context.Context, owner, repo, ref string) ([]CheckRun, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/commits/%s/check-runs", c.apiURL, owner, repo, ref)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.addHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list check runs, status: %d", resp.StatusCode)
	}

	var result struct {
		CheckRuns []CheckRun `json:"check_runs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.CheckRuns, nil
}

func (m *MockClient) ListCheckRunsForRef(ctx context.Context, owner, repo, ref string) ([]CheckRun, error) {
	if m.ListCheckRunsForRefFunc != nil {
		return m.ListCheckRunsForRefFunc(ctx, owner, repo, ref)
	}
	return nil, nil
}

func (c *RESTClient) CreatePR(ctx context.Context, owner, repo, title, branch, body string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls", c.apiURL, owner, repo)

	payload := map[string]interface{}{
		"title": title,
		"head":  branch,
		"base":  "main",
		"body":  body,
	}

	payloadBytes, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", err
	}
	c.addHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create pr failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.HTMLURL, nil
}

func (c *RESTClient) CommitAndPushFile(ctx context.Context, owner, repo, branch, filePath, fileContent, commitMessage string) error {
	mainSHA, err := c.getBranchSHA(ctx, owner, repo, "main")
	if err != nil {
		return err
	}

	if err := c.createBranch(ctx, owner, repo, branch, mainSHA); err != nil {
		return err
	}

	blobURL := fmt.Sprintf("%s/repos/%s/%s/git/blobs", c.apiURL, owner, repo)
	blobPayload := map[string]string{
		"content":  fileContent,
		"encoding": "utf-8",
	}
	blobBytes, _ := json.Marshal(blobPayload)
	reqBlob, err := http.NewRequestWithContext(ctx, http.MethodPost, blobURL, bytes.NewBuffer(blobBytes))
	if err != nil {
		return err
	}
	c.addHeaders(reqBlob)

	respBlob, err := c.client.Do(reqBlob)
	if err != nil {
		return err
	}
	defer respBlob.Body.Close()

	if respBlob.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(respBlob.Body)
		return fmt.Errorf("create blob failed with status: %d, body: %s", respBlob.StatusCode, string(bodyBytes))
	}

	var blobResult struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(respBlob.Body).Decode(&blobResult); err != nil {
		return err
	}

	treeURL := fmt.Sprintf("%s/repos/%s/%s/git/trees", c.apiURL, owner, repo)
	treePayload := map[string]interface{}{
		"base_tree": mainSHA,
		"tree": []map[string]interface{}{
			{
				"path": filePath,
				"mode": "100644",
				"type": "blob",
				"sha":  blobResult.SHA,
			},
		},
	}
	treeBytes, _ := json.Marshal(treePayload)
	reqTree, err := http.NewRequestWithContext(ctx, http.MethodPost, treeURL, bytes.NewBuffer(treeBytes))
	if err != nil {
		return err
	}
	c.addHeaders(reqTree)

	respTree, err := c.client.Do(reqTree)
	if err != nil {
		return err
	}
	defer respTree.Body.Close()

	if respTree.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(respTree.Body)
		return fmt.Errorf("create tree failed with status: %d, body: %s", respTree.StatusCode, string(bodyBytes))
	}

	var treeResult struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(respTree.Body).Decode(&treeResult); err != nil {
		return err
	}

	commitURL := fmt.Sprintf("%s/repos/%s/%s/git/commits", c.apiURL, owner, repo)
	commitPayload := map[string]interface{}{
		"message": commitMessage,
		"tree":    treeResult.SHA,
		"parents": []string{mainSHA},
	}
	commitBytes, _ := json.Marshal(commitPayload)
	reqCommit, err := http.NewRequestWithContext(ctx, http.MethodPost, commitURL, bytes.NewBuffer(commitBytes))
	if err != nil {
		return err
	}
	c.addHeaders(reqCommit)

	respCommit, err := c.client.Do(reqCommit)
	if err != nil {
		return err
	}
	defer respCommit.Body.Close()

	if respCommit.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(respCommit.Body)
		return fmt.Errorf("create commit failed with status: %d, body: %s", respCommit.StatusCode, string(bodyBytes))
	}

	var commitResult struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(respCommit.Body).Decode(&commitResult); err != nil {
		return err
	}

	refURL := fmt.Sprintf("%s/repos/%s/%s/git/refs/heads/%s", c.apiURL, owner, repo, branch)
	refPayload := map[string]interface{}{
		"sha":   commitResult.SHA,
		"force": true,
	}
	refBytes, _ := json.Marshal(refPayload)
	reqRef, err := http.NewRequestWithContext(ctx, http.MethodPatch, refURL, bytes.NewBuffer(refBytes))
	if err != nil {
		return err
	}
	c.addHeaders(reqRef)

	respRef, err := c.client.Do(reqRef)
	if err != nil {
		return err
	}
	defer respRef.Body.Close()

	if respRef.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(respRef.Body)
		return fmt.Errorf("update ref failed with status: %d, body: %s", respRef.StatusCode, string(bodyBytes))
	}

	return nil
}

func (m *MockClient) CreatePR(ctx context.Context, owner, repo, title, branch, body string) (string, error) {
	return "https://github.com/mock/mock/pull/2", nil
}

func (m *MockClient) CommitAndPushFile(ctx context.Context, owner, repo, branch, filePath, fileContent, commitMessage string) error {
	return nil
}
