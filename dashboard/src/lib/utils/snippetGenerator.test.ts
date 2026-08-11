// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

import { describe, it, expect } from 'vitest';
import { generateSnippets } from './snippetGenerator';

describe('snippetGenerator', () => {
	it('generates a basic GET request', () => {
		const snippets = generateSnippets('GET', 'https://api.example.com/users');

		expect(snippets.curl).toContain('curl -X GET "https://api.example.com/users"');
		expect(snippets.python).toContain('requests.get(url)');
		expect(snippets.node).toContain("fetch('https://api.example.com/users'");
		expect(snippets.node).toContain("method: 'GET'");
		expect(snippets.go).toContain('http.NewRequest("GET", url, nil)');
	});

	it('generates a POST request with headers and JSON body', () => {
		const headers = {
			'Content-Type': 'application/json',
			'Authorization': 'Bearer token123'
		};
		const body = { name: 'John Doe', age: 30 };
		const snippets = generateSnippets('POST', 'https://api.example.com/users', headers, {}, body);

		// cURL
		expect(snippets.curl).toContain('curl -X POST "https://api.example.com/users"');
		expect(snippets.curl).toContain('-H "Content-Type: application/json"');
		expect(snippets.curl).toContain('-H "Authorization: Bearer token123"');
		expect(snippets.curl).toContain(`-d '{"name":"John Doe","age":30}'`);

		// Python
		expect(snippets.python).toContain('headers = {');
		expect(snippets.python).toContain('"Content-Type": "application/json"');
		expect(snippets.python).toContain('json_data = {');
		expect(snippets.python).toContain('requests.post(url, headers=headers, json=json_data)');

		// Node
		expect(snippets.node).toContain("method: 'POST'");
		expect(snippets.node).toContain("'Content-Type': 'application/json'");
		expect(snippets.node).toContain(`body: JSON.stringify({"name":"John Doe","age":30})`);

		// Go
		expect(snippets.go).toContain("payload := strings.NewReader(`{\"name\":\"John Doe\",\"age\":30}`)");
		expect(snippets.go).toContain(`req, _ := http.NewRequest("POST", url, payload)`);
		expect(snippets.go).toContain(`req.Header.Add("Content-Type", "application/json")`);
	});

	it('generates request with query parameters', () => {
		const queryParams = {
			page: '1',
			limit: '10'
		};
		const snippets = generateSnippets('GET', 'https://api.example.com/users', {}, queryParams);

		const expectedUrl = 'https://api.example.com/users?page=1&limit=10';
		expect(snippets.curl).toContain(`"${expectedUrl}"`);
		expect(snippets.python).toContain(`url = "${expectedUrl}"`);
		expect(snippets.node).toContain(`fetch('${expectedUrl}'`);
		expect(snippets.go).toContain(`url := "${expectedUrl}"`);
	});

	it('escapes quotes appropriately', () => {
		const body = { message: "It's a beautiful day" };
		const snippets = generateSnippets('POST', 'https://api.example.com/comments', {}, {}, body);

		// cURL single quote escaping: "It's" -> "It'\''s"
		expect(snippets.curl).toContain(`-d '{"message":"It'\\''s a beautiful day"}'`);

		// Go raw string literals with backticks
		expect(snippets.go).toContain(`strings.NewReader(\`{"message":"It's a beautiful day"}\`)`);
	});

	it('handles boolean and null values in Python snippets', () => {
		const body = { isActive: true, hasError: false, metadata: null };
		const snippets = generateSnippets('POST', 'https://api.example.com/status', {}, {}, body);

		expect(snippets.python).toContain(': True');
		expect(snippets.python).toContain(': False');
		expect(snippets.python).toContain(': None');
	});
});
