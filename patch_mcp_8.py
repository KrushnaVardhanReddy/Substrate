import sys

filename = 'engine/cmd/substrate-mcp/main.go'
with open(filename, 'r') as f:
    content = f.read()


tool3_search = """			registryURL := os.Getenv("REGISTRY_API_URL")
			if registryURL == "" {
				registryURL = "http://localhost:8090"
			}
			url := fmt.Sprintf("%s/api/v1/history/%s/%s?limit=%d", registryURL, org, repo, limit)

			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create request: %w", err)
			}

			token := os.Getenv("REGISTRY_API_TOKEN")
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			if runtime.GOOS == "wasip1" {
				return nil, fmt.Errorf("WASI does not support network requests")
			}
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch history: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("registry API returned status %d", resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			return string(body), nil"""

tool3_replace = """			var body []byte

			// Try local cache first
			if cache.GlobalCache != nil {
				body, _ = cache.GlobalCache.GetBreakingChangeHistory(input.Repo)
			}

			if body == nil {
				registryURL := os.Getenv("REGISTRY_API_URL")
				if registryURL == "" {
					registryURL = "http://localhost:8090"
				}
				url := fmt.Sprintf("%s/api/v1/history/%s/%s?limit=%d", registryURL, org, repo, limit)

				req, err := http.NewRequest(http.MethodGet, url, nil)
				if err != nil {
					return nil, fmt.Errorf("failed to create request: %w", err)
				}

				token := os.Getenv("REGISTRY_API_TOKEN")
				if token != "" {
					req.Header.Set("Authorization", "Bearer "+token)
				}

				if runtime.GOOS == "wasip1" {
					return nil, fmt.Errorf("WASI does not support network requests")
				}
				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch history: %w", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					return nil, fmt.Errorf("registry API returned status %d", resp.StatusCode)
				}

				body, err = io.ReadAll(resp.Body)
				if err != nil {
					return nil, err
				}
			}

			return string(body), nil"""

content = content.replace(tool3_search, tool3_replace)

with open(filename, 'w') as f:
    f.write(content)
