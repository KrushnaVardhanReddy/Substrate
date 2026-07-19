import sys

filename = 'engine/cmd/substrate-mcp/main.go'
with open(filename, 'r') as f:
    content = f.read()

# Make sure Tool 3 `get_breaking_change_history` uses cache.GlobalCache.GetBreakingChangeHistory
# It looks like my previous regex search didn't find Tool 3 because the search string wasn't exactly right.
tool3_search = """			apiURL := os.Getenv("REGISTRY_API_URL")
			if apiURL == "" {
				apiURL = "http://localhost:8090"
			}

			url := fmt.Sprintf("%s/api/v1/history/%s/%s?limit=%d", apiURL, org, repo, limit)
			if runtime.GOOS == "wasip1" {
				return nil, fmt.Errorf("WASI does not support network requests")
			}
			resp, err := http.Get(url)
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

tool3_replace = """			apiURL := os.Getenv("REGISTRY_API_URL")
			if apiURL == "" {
				apiURL = "http://localhost:8090"
			}

			var body []byte

			// Try local cache first
			if cache.GlobalCache != nil {
				body, _ = cache.GlobalCache.GetBreakingChangeHistory(input.Repo)
			}

			if body == nil {
				url := fmt.Sprintf("%s/api/v1/history/%s/%s?limit=%d", apiURL, org, repo, limit)
				if runtime.GOOS == "wasip1" {
					return nil, fmt.Errorf("WASI does not support network requests")
				}
				resp, err := http.Get(url)
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

if tool3_search in content:
    content = content.replace(tool3_search, tool3_replace)
    with open(filename, 'w') as f:
        f.write(content)
else:
    print("WARNING: Tool 3 search string not found!")
