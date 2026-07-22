package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

// For testing purposes
var githubApiBaseUrl = "https://api.github.com"

type EnforceRequest struct {
	Enforce bool `json:"enforce"`
}

type RepoInfo struct {
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch"`
}

func handleGitHubRateLimit(resp *http.Response) {
	if resp == nil {
		return
	}
	remainingStr := resp.Header.Get("X-RateLimit-Remaining")
	if remainingStr == "" {
		return
	}

	remaining, err := strconv.Atoi(remainingStr)
	if err != nil || remaining > 5 {
		if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusTooManyRequests {
			return
		}
	}

	resetStr := resp.Header.Get("X-RateLimit-Reset")
	if resetStr == "" {
		time.Sleep(10 * time.Second)
		return
	}

	resetTimeUnix, err := strconv.ParseInt(resetStr, 10, 64)
	if err != nil {
		time.Sleep(10 * time.Second)
		return
	}

	resetTime := time.Unix(resetTimeUnix, 0)
	now := time.Now()

	if resetTime.After(now) {
		sleepDuration := resetTime.Sub(now) + time.Second
		fmt.Printf("GitHub API rate limit approaching or hit. Sleeping for %v until %v\n", sleepDuration, resetTime)
		time.Sleep(sleepDuration)
	}
}

func EnforceGlobalHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org := chi.URLParam(r, "org")
		if org == "" {
			http.Error(w, "missing org", http.StatusBadRequest)
			return
		}

		var req EnforceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		authHeader := r.Header.Get("Authorization")
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			http.Error(w, "missing Authorization token", http.StatusUnauthorized)
			return
		}

		if err := updateSubstrateGlobalEnforcement(token, org, req.Enforce); err != nil {
			http.Error(w, fmt.Sprintf("failed to update global enforcement: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}
}

func updateSubstrateGlobalEnforcement(token, org string, enforce bool) error {
	client := &http.Client{}

	page := 1
	for {
		url := fmt.Sprintf("%s/orgs/%s/repos?per_page=100&page=%d", githubApiBaseUrl, org, page)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		var resp *http.Response
		var retryCount int
		for retryCount < 3 {
			resp, err = client.Do(req)
			if err != nil {
				return err
			}
			handleGitHubRateLimit(resp)

			if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
				resp.Body.Close()
				retryCount++
				continue
			}
			break
		}

		if resp == nil {
			return fmt.Errorf("failed to execute request after retries")
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return fmt.Errorf("failed to list repos: %d %s", resp.StatusCode, string(body))
		}

		var repos []RepoInfo
		if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
			resp.Body.Close()
			return err
		}
		resp.Body.Close()

		if len(repos) == 0 {
			break
		}

		for _, repo := range repos {
			if repo.DefaultBranch == "" {
				repo.DefaultBranch = "main"
			}

			if enforce {
				protectUrl := fmt.Sprintf("%s/repos/%s/%s/branches/%s/protection", githubApiBaseUrl, org, repo.Name, repo.DefaultBranch)
				payload := map[string]interface{}{
					"required_status_checks": map[string]interface{}{
						"strict":   true,
						"contexts": []string{"substrate"},
					},
					"enforce_admins":                false,
					"required_pull_request_reviews": nil,
					"restrictions":                  nil,
				}

				payloadBytes, err := json.Marshal(payload)
				if err != nil {
					continue
				}

				putReq, err := http.NewRequest("PUT", protectUrl, bytes.NewReader(payloadBytes))
				if err != nil {
					continue
				}
				putReq.Header.Set("Authorization", "Bearer "+token)
				putReq.Header.Set("Accept", "application/vnd.github.v3+json")
				putReq.Header.Set("Content-Type", "application/json")

				var putResp *http.Response
				var putRetryCount int
				for putRetryCount < 3 {
					putResp, err = client.Do(putReq)
					if err != nil {
						break
					}
					handleGitHubRateLimit(putResp)

					if putResp.StatusCode == http.StatusForbidden || putResp.StatusCode == http.StatusTooManyRequests {
						putResp.Body.Close()
						putRetryCount++
						continue
					}
					break
				}
				if putResp != nil {
					putResp.Body.Close()
				}
			} else {
				// Remove the context
				contextsUrl := fmt.Sprintf("%s/repos/%s/%s/branches/%s/protection/required_status_checks/contexts", githubApiBaseUrl, org, repo.Name, repo.DefaultBranch)

				// DELETE request payload requires an array of strings
				payload := []string{"substrate"}
				payloadBytes, _ := json.Marshal(payload)

				delReq, err := http.NewRequest("DELETE", contextsUrl, bytes.NewReader(payloadBytes))
				if err != nil {
					continue
				}
				delReq.Header.Set("Authorization", "Bearer "+token)
				delReq.Header.Set("Accept", "application/vnd.github.v3+json")
				delReq.Header.Set("Content-Type", "application/json")

				var delResp *http.Response
				var delRetryCount int
				for delRetryCount < 3 {
					delResp, err = client.Do(delReq)
					if err != nil {
						break
					}
					handleGitHubRateLimit(delResp)

					if delResp.StatusCode == http.StatusForbidden || delResp.StatusCode == http.StatusTooManyRequests {
						delResp.Body.Close()
						delRetryCount++
						continue
					}
					break
				}
				if delResp != nil {
					delResp.Body.Close()
				}
			}
		}

		page++
	}

	return nil
}
