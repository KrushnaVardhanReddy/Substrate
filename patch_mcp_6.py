import sys

filename = 'engine/cmd/substrate-mcp/main.go'
with open(filename, 'r') as f:
    content = f.read()

# Update check_compatibility tool
# Currently it creates an empty base file. Let's make it fetch from cache/API
tool2_search = """			// Write empty base schema to temp file
			baseFile, err := os.CreateTemp("", "base-*.schema")
			if err != nil {
				return nil, fmt.Errorf("failed to create temp base file: %v", err)
			}
			defer os.Remove(baseFile.Name())
			// Initialize with an empty valid structure for some formats
			baseContent := ""
			if args.SchemaType == "openapi" {
				baseContent = `{"openapi":"3.0.0","info":{"title":"mock","version":"1"},"paths":{}}`
			}
			if _, err := baseFile.WriteString(baseContent); err != nil {
				return nil, err
			}
			baseFile.Close()"""

tool2_replace = """			var baseSchemaContent []byte

			// Try local cache first
			if cache.GlobalCache != nil {
				baseSchemaContent, _ = cache.GlobalCache.GetSchema(args.ProviderRepo)
			}

			if baseSchemaContent == nil {
				apiURL := os.Getenv("REGISTRY_API_URL")
				if apiURL == "" {
					apiURL = "http://localhost:8090"
				}
				parts := strings.SplitN(args.ProviderRepo, "/", 2)
				if len(parts) == 2 {
					url := fmt.Sprintf("%s/api/v1/schema/%s/%s", apiURL, parts[0], parts[1])
					if runtime.GOOS != "wasip1" {
						resp, err := http.Get(url)
						if err == nil {
							defer resp.Body.Close()
							if resp.StatusCode == http.StatusOK {
								baseSchemaContent, _ = io.ReadAll(resp.Body)
							}
						}
					}
				}
			}

			// Write empty base schema to temp file if not found
			baseFile, err := os.CreateTemp("", "base-*.schema")
			if err != nil {
				return nil, fmt.Errorf("failed to create temp base file: %v", err)
			}
			defer os.Remove(baseFile.Name())

			if len(baseSchemaContent) > 0 {
				if _, err := baseFile.Write(baseSchemaContent); err != nil {
					return nil, err
				}
			} else {
				// Initialize with an empty valid structure for some formats
				baseContent := ""
				if args.SchemaType == "openapi" {
					baseContent = `{"openapi":"3.0.0","info":{"title":"mock","version":"1"},"paths":{}}`
				}
				if _, err := baseFile.WriteString(baseContent); err != nil {
					return nil, err
				}
			}
			baseFile.Close()"""

content = content.replace(tool2_search, tool2_replace)

with open(filename, 'w') as f:
    f.write(content)
