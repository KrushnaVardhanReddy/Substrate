import re

with open('engine/cmd/substrate-mcp/main.go', 'r') as f:
    content = f.read()

content = content.replace('resp, err := http.Get(fmt.Sprintf("%s/api/v1/graph/%s", apiURL, args.Org))',
"""
            if os.Getenv("GOOS") == "wasip1" || os.Getenv("WASMTIME") == "1" {
                return nil, fmt.Errorf("WASI does not support network requests")
            }
            resp, err := http.Get(fmt.Sprintf("%s/api/v1/graph/%s", apiURL, args.Org))""")
with open('engine/cmd/substrate-mcp/main.go', 'w') as f:
    f.write(content)
