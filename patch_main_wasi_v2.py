import re

with open('engine/cmd/substrate-mcp/main.go', 'r') as f:
    content = f.read()

replacements = [
    (
        'resp, err := http.Get(fmt.Sprintf("%s/api/v1/graph/%s", apiURL, args.Org))',
        '''if runtime.GOOS == "wasip1" {
				return nil, fmt.Errorf("WASI does not support network requests")
			}
			resp, err := http.Get(fmt.Sprintf("%s/api/v1/graph/%s", apiURL, args.Org))'''
    ),
    (
        'resp, err := http.DefaultClient.Do(req)',
        '''if runtime.GOOS == "wasip1" {
				return nil, fmt.Errorf("WASI does not support network requests")
			}
			resp, err := http.DefaultClient.Do(req)'''
    ),
    (
        'resp, err := http.Get(url)',
        '''if runtime.GOOS == "wasip1" {
				return nil, fmt.Errorf("WASI does not support network requests")
			}
			resp, err := http.Get(url)'''
    )
]

for old, new in replacements:
    content = content.replace(old, new)

with open('engine/cmd/substrate-mcp/main.go', 'w') as f:
    f.write(content)
