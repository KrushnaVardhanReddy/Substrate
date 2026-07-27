package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	ForgejoURL = "http://127.0.0.1:3000"
)

func SetupForgejo(orgName, repoName string) (string, string, error) {
	user := "adminuser"
	password := "Admin123!"

	installForm := url.Values{
		"db_type":                           {"sqlite3"},
		"db_host":                           {"localhost:3306"},
		"db_user":                           {"root"},
		"db_passwd":                         {""},
		"db_name":                           {"gitea"},
		"ssl_mode":                          {"disable"},
		"db_path":                           {"/data/gitea/gitea.db"},
		"app_name":                          {"Forgejo E2E"},
		"repo_root_path":                    {"/data/git/repositories"},
		"run_user":                          {"git"},
		"domain":                            {"localhost"},
		"ssh_port":                          {"22"},
		"http_port":                         {"3000"},
		"app_url":                           {"http://localhost:3000/"},
		"log_root_path":                     {"/data/gitea/log"},
		"enable_update_checker":             {"on"},
		"admin_name":                        {user},
		"admin_passwd":                    {password},
		"admin_confirm_passwd":            {password},
		"admin_email":                       {"admin@example.com"},
	}

	installReq, _ := http.NewRequest("POST", ForgejoURL+"/", strings.NewReader(installForm.Encode()))
	installReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	installResp, err := client.Do(installReq)
	if err == nil {
		if installResp.StatusCode != 200 && installResp.StatusCode != 302 {
			fmt.Printf("Install Form returned %d\n", installResp.StatusCode)
		}
		installResp.Body.Close()
	}

	// Ensure Gitea is up by checking /api/v1/version
	for i := 0; i < 30; i++ {
		resp, err := http.Get("http://127.0.0.1:3000/api/v1/version")
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
	req, _ = http.NewRequest("POST", ForgejoURL+"/api/v1/orgs/"+orgName+"/repos", bytes.NewBuffer(repoBytes))
	req.SetBasicAuth(user, password)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
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
		"type": "github",
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

	cloneURL := fmt.Sprintf("http://%s:%s@127.0.0.1:3000/%s/%s.git", user, password, orgName, repoName)
	return cloneURL, password, nil
}
