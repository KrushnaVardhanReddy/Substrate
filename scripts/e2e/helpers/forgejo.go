package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"
)

const (
	ForgejoURL = "http://127.0.0.1:3005"
)

func SetupForgejo(orgName, repoName string) (string, string, error) {
	user := "adminuser"
	password := "Admin123!"
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Timeout: 30 * time.Second,
		Jar:     jar,
	}

	// Ensure Gitea is up by checking /api/v1/version
	for i := 0; i < 30; i++ {
		resp, err := client.Get(ForgejoURL + "/api/v1/version")
		if err == nil {
			if resp.StatusCode == 200 {
				resp.Body.Close()
				break
			}
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	time.Sleep(2 * time.Second)

	orgPayload := map[string]interface{}{
		"username":   orgName,
		"visibility": "public",
	}
	orgBytes, _ := json.Marshal(orgPayload)
	req, _ := http.NewRequest("POST", ForgejoURL+"/api/v1/orgs", bytes.NewBuffer(orgBytes))
	req.SetBasicAuth(user, password)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to create org: %v", err)
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusUnprocessableEntity {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return "", "", fmt.Errorf("failed to create org, status %d: %s", resp.StatusCode, string(body))
	}
	resp.Body.Close()

	repoPayload := map[string]interface{}{
		"name":      repoName,
		"private":   false,
		"auto_init": true,
	}
	repoBytes, _ := json.Marshal(repoPayload)
	req2, _ := http.NewRequest("POST", ForgejoURL+"/api/v1/orgs/"+orgName+"/repos", bytes.NewBuffer(repoBytes))
	req2.SetBasicAuth(user, password)
	req2.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req2)
	if err != nil {
		return "", "", fmt.Errorf("failed to create repo: %v", err)
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return "", "", fmt.Errorf("failed to create repo, status %d: %s", resp.StatusCode, string(body))
	}
	resp.Body.Close()

	time.Sleep(2 * time.Second)

	webhookPayload := map[string]interface{}{
		"type": "gitea",
		"config": map[string]string{
			"url":          "http://127.0.0.1:8090/api/v1/webhook",
			"content_type": "json",
			"secret":       "test-secret",
		},
		"events": []string{"push"},
		"active": true,
	}
	webhookBytes, _ := json.Marshal(webhookPayload)
	req, _ = http.NewRequest("POST", ForgejoURL+"/api/v1/repos/"+orgName+"/"+repoName+"/hooks", bytes.NewBuffer(webhookBytes))
	req.SetBasicAuth(user, password)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to create webhook: %v", err)
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Webhook creation returned: %d %s\n", resp.StatusCode, string(body))
	}
	resp.Body.Close()

	cloneURL := fmt.Sprintf("http://%s:%s@127.0.0.1:3005/%s/%s.git", user, password, orgName, repoName)
	return cloneURL, password, nil
}
