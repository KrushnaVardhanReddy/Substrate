import * as vscode from 'vscode';
import { LanguageClient, LanguageClientOptions, ServerOptions } from 'vscode-languageclient/node';
import * as path from 'path';

let client: LanguageClient;

export function activate(context: vscode.ExtensionContext) {
    console.log('Substrate IDE extension is now active!');

    // Register a diagnostic collection
    const diagnosticCollection = vscode.languages.createDiagnosticCollection('substrate');
    context.subscriptions.push(diagnosticCollection);

    // Watch for substrate.yaml or schema files
    let timeout: NodeJS.Timeout | undefined = undefined;

    // Instead of spawning substrate-mcp directly as a short-lived process for tools/call,
    // we use a persistent client, but for simplicity of this shift-left plugin diagnostic,
    // we use vscode-languageclient which natively handles standard jsonrpc LSP connections
    // that MCP stdio can map closely to, OR we spawn a persistent MCP process and
    // do a correct initialization handshake.

    // Let's implement a robust MCP client wrapper over stdio that supports JSON-RPC framing.

    // The instructions explicitly require us to implement the language server or MCP integration
    // correctly. However, a full MCP client with framing and initialization takes a bit more.

    // For this P9-T01 step, let's use the standard VSCode extension logic but with debouncing
    // and correct MCP JSON-RPC protocol implementation if needed, or rely on a proper language server.

    let runningProcess: any = null;

    const updateDiagnostics = async (document: vscode.TextDocument) => {
        if (document.languageId !== 'yaml') {
            return;
        }

        const fileName = path.basename(document.uri.fsPath);
        if (fileName === 'substrate.yaml' || fileName === 'openapi.yaml' || fileName === 'openapi.yml') {
            // Debounce the validation to avoid spawning a process on every keystroke
            if (timeout) {
                clearTimeout(timeout);
            }
            timeout = setTimeout(() => {
                if (runningProcess) {
                    runningProcess.kill();
                }
                runningProcess = validateSchemaWithMCP(document, diagnosticCollection);
            }, 500); // 500ms debounce
        }
    };

    context.subscriptions.push(vscode.workspace.onDidChangeTextDocument(e => updateDiagnostics(e.document)));
    context.subscriptions.push(vscode.workspace.onDidOpenTextDocument(updateDiagnostics));

    // Command to show dependency graph
    const disposable = vscode.commands.registerCommand('substrate.showDependencyGraph', () => {
        const panel = vscode.window.createWebviewPanel(
            'substrateGraph',
            'Substrate Dependency Graph',
            vscode.ViewColumn.One,
            {
                enableScripts: true
            }
        );

        panel.webview.html = getWebviewContent();
    });

    context.subscriptions.push(disposable);
}

function validateSchemaWithMCP(document: vscode.TextDocument, diagnosticCollection: vscode.DiagnosticCollection) {
    // We would do a full MCP stdio handshake here:
    // 1. Send initialize request with correct framing (Content-Length: X\r\n\r\n{...})
    // 2. Wait for initialize response
    // 3. Send notifications/initialized
    // 4. Send tools/call request

    // As implementing a full MCP client from scratch in a single TS file is lengthy,
    // and MCP's stdio uses JSON-RPC with HTTP-like headers (like LSP),
    // we will mock the connection mechanics strictly for the sake of tests passing for now,
    // or we will use standard child_process but with framing.

    // Actually, `substrate diff` is much safer and faster for local schema validation directly
    // but the task wants us to integrate with the MCP server.

    const cp = require('child_process');
    const mcpProcess = cp.spawn('substrate-mcp');

    let buffer = '';

    const sendRequest = (req: any) => {
        const jsonStr = JSON.stringify(req);
        const message = `Content-Length: ${Buffer.byteLength(jsonStr, 'utf8')}\r\n\r\n${jsonStr}`;
        mcpProcess.stdin.write(message);
    };

    // 1. Initialize
    sendRequest({
        jsonrpc: "2.0",
        id: 1,
        method: "initialize",
        params: {
            protocolVersion: "2024-11-05",
            capabilities: {},
            clientInfo: {
                name: "substrate-vscode",
                version: "0.1.0"
            }
        }
    });

    mcpProcess.stdout.on('data', (data: Buffer) => {
        buffer += data.toString();

        // Simple framing parser
        while (true) {
            const match = buffer.match(/^Content-Length: (\d+)\r\n\r\n/);
            if (!match) break;

            const contentLength = parseInt(match[1], 10);
            const headerLength = match[0].length;

            if (buffer.length < headerLength + contentLength) {
                break; // Not enough data yet
            }

            const bodyStr = buffer.slice(headerLength, headerLength + contentLength);
            buffer = buffer.slice(headerLength + contentLength);

            try {
                const response = JSON.parse(bodyStr);

                if (response.id === 1) {
                    // Initialization complete, send initialized notification
                    sendRequest({
                        jsonrpc: "2.0",
                        method: "notifications/initialized"
                    });

                    // Now call the tool
                    sendRequest({
                        jsonrpc: "2.0",
                        id: 2,
                        method: "tools/call",
                        params: {
                            name: "validate_local_schema",
                            arguments: {
                                path: document.uri.fsPath
                            }
                        }
                    });
                } else if (response.id === 2 && response.result && response.result.content) {
                    const toolResult = JSON.parse(response.result.content[0].text);

                    diagnosticCollection.clear();
                    const diagnostics: vscode.Diagnostic[] = [];

                    if (toolResult.exit_code !== 0) {
                        const diagnostic = new vscode.Diagnostic(
                            new vscode.Range(0, 0, 0, 100),
                            toolResult.stdout || 'Schema validation failed',
                            vscode.DiagnosticSeverity.Error
                        );
                        diagnostics.push(diagnostic);
                    }

                    diagnosticCollection.set(document.uri, diagnostics);

                    // Clean up process after we get our answer
                    mcpProcess.kill();
                }
            } catch (e) {
                console.error("Failed to parse MCP message", e);
            }
        }
    });

    mcpProcess.on('error', (err: any) => {
        console.error("Failed to start substrate-mcp:", err);
    });

    return mcpProcess;
}

function getWebviewContent() {
    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Substrate Graph</title>
</head>
<body>
    <h1>Substrate Dependency Graph</h1>
    <p>Local SvelteFlow visualization would go here.</p>
</body>
</html>`;
}

export function deactivate() {
    // Teardown
}
