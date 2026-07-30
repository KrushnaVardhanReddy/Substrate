export function generateSnippets(
	method: string,
	path: string,
	headers: Record<string, string> = {},
	queryParams: Record<string, string> = {},
	body: any = null
): Record<'curl' | 'python' | 'node' | 'go', string> {
	const methodUpper = method.toUpperCase();
	let queryString = '';
	if (Object.keys(queryParams).length > 0) {
		const params = new URLSearchParams();
		for (const [key, value] of Object.entries(queryParams)) {
			params.append(key, value);
		}
		queryString = `?${params.toString()}`;
	}
	const fullUrl = `${path}${queryString}`;

	return {
		curl: generateCurlSnippet(methodUpper, fullUrl, headers, body),
		python: generatePythonSnippet(methodUpper, fullUrl, headers, body),
		node: generateNodeSnippet(methodUpper, fullUrl, headers, body),
		go: generateGoSnippet(methodUpper, fullUrl, headers, body)
	};
}

function generateCurlSnippet(
	method: string,
	url: string,
	headers: Record<string, string>,
	body: any
): string {
	let snippet = `curl -X ${method} "${url}"`;
	for (const [key, value] of Object.entries(headers)) {
		snippet += ` \\\n  -H "${key}: ${value}"`;
	}
	if (body) {
		const bodyStr = typeof body === 'string' ? body : JSON.stringify(body);
		// Escape single quotes for bash
		const escapedBody = bodyStr.replace(/'/g, "'\\''");
		snippet += ` \\\n  -d '${escapedBody}'`;
	}
	return snippet;
}

function generatePythonSnippet(
	method: string,
	url: string,
	headers: Record<string, string>,
	body: any
): string {
	let snippet = `import requests\n\n`;
	snippet += `url = "${url}"\n\n`;

	if (Object.keys(headers).length > 0) {
		snippet += `headers = ${JSON.stringify(headers, null, 4)}\n\n`;
	}

	if (body) {
		if (typeof body === 'object') {
			// Convert JSON literal representation to Python dict representation
			let pythonDict = JSON.stringify(body, null, 4);
			pythonDict = pythonDict.replace(/: true/g, ': True');
			pythonDict = pythonDict.replace(/: false/g, ': False');
			pythonDict = pythonDict.replace(/: null/g, ': None');
			snippet += `json_data = ${pythonDict}\n\n`;
		} else {
			let pythonData = JSON.stringify(body);
			if (pythonData === 'true') pythonData = 'True';
			if (pythonData === 'false') pythonData = 'False';
			if (pythonData === 'null') pythonData = 'None';
			snippet += `data = ${pythonData}\n\n`;
		}
	}

	snippet += `response = requests.${method.toLowerCase()}(url`;
	if (Object.keys(headers).length > 0) {
		snippet += `, headers=headers`;
	}
	if (body) {
		if (typeof body === 'object') {
			snippet += `, json=json_data`;
		} else {
			snippet += `, data=data`;
		}
	}
	snippet += `)\n`;
	snippet += `print(response.json())`;
	return snippet;
}

function generateNodeSnippet(
	method: string,
	url: string,
	headers: Record<string, string>,
	body: any
): string {
	let snippet = `fetch('${url}', {\n`;
	snippet += `  method: '${method}',\n`;

	if (Object.keys(headers).length > 0) {
		snippet += `  headers: {\n`;
		for (const [key, value] of Object.entries(headers)) {
			snippet += `    '${key}': '${value}',\n`;
		}
		snippet += `  },\n`;
	}

	if (body) {
		const bodyStr = typeof body === 'string' ? body : JSON.stringify(body);
		snippet += `  body: JSON.stringify(${bodyStr})\n`;
	}

	snippet += `})\n`;
	snippet += `  .then(response => response.json())\n`;
	snippet += `  .then(data => console.log(data))\n`;
	snippet += `  .catch(error => console.error('Error:', error));`;

	return snippet;
}

function generateGoSnippet(
	method: string,
	url: string,
	headers: Record<string, string>,
	body: any
): string {
	let snippet = `package main\n\n`;
	snippet += `import (\n`;
	snippet += `\t"fmt"\n`;
	snippet += `\t"io"\n`;
	snippet += `\t"net/http"\n`;
	if (body) {
		snippet += `\t"strings"\n`;
	}
	snippet += `)\n\n`;
	snippet += `func main() {\n`;
	snippet += `\turl := "${url}"\n`;

	if (body) {
		const bodyStr = typeof body === 'string' ? body : JSON.stringify(body);
		// Use backticks for raw string literal to support multiline and avoid escaping double quotes
		snippet += `\tpayload := strings.NewReader(\`${bodyStr}\`)\n`;
		snippet += `\treq, _ := http.NewRequest("${method}", url, payload)\n`;
	} else {
		snippet += `\treq, _ := http.NewRequest("${method}", url, nil)\n`;
	}

	for (const [key, value] of Object.entries(headers)) {
		snippet += `\treq.Header.Add("${key}", "${value}")\n`;
	}

	snippet += `\n\tres, err := http.DefaultClient.Do(req)\n`;
	snippet += `\tif err != nil {\n`;
	snippet += `\t\tfmt.Println(err)\n`;
	snippet += `\t\treturn\n`;
	snippet += `\t}\n`;
	snippet += `\tdefer res.Body.Close()\n\n`;
	snippet += `\tbody, _ := io.ReadAll(res.Body)\n`;
	snippet += `\tfmt.Println(string(body))\n`;
	snippet += `}\n`;

	return snippet;
}
