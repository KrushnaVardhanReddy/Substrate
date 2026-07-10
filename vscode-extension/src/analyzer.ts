import * as vscode from 'vscode';
import * as cp from 'child_process';
import * as util from 'util';
import * as path from 'path';
import * as os from 'os';
import * as fs from 'fs';

const execFile = util.promisify(cp.execFile);

export async function analyzeDocument(document: vscode.TextDocument, diagnosticCollection: vscode.DiagnosticCollection) {
    if (document.uri.scheme !== 'file') {
        return;
    }

    const filePath = document.uri.fsPath;
    const workspaceFolder = vscode.workspace.getWorkspaceFolder(document.uri);
    const cwd = workspaceFolder ? workspaceFolder.uri.fsPath : path.dirname(filePath);

    // Get the relative path for git
    const relativePath = workspaceFolder ? path.relative(cwd, filePath) : path.basename(filePath);

    let tempFilePath = '';
    try {
        // Get the previous commit file using git show HEAD:<file>
        const { stdout: gitOutput } = await execFile('git', ['show', `HEAD:${relativePath}`], { cwd });

        // Write it to a temp file
        tempFilePath = path.join(os.tmpdir(), `substrate_prev_${Date.now()}_${path.basename(filePath)}`);
        fs.writeFileSync(tempFilePath, gitOutput);

        // Run substrate diff
        const { stdout, stderr } = await execFile('substrate', ['diff', tempFilePath, filePath, '--format', 'json'], { cwd });

        // Output might have non-json stuff before or after, but assuming it's valid JSON for now.
        processOutput(stdout, document, diagnosticCollection);

    } catch (error: any) {
        // If substrate exits with non-zero, the output is in error.stdout
        if (error.stdout) {
            processOutput(error.stdout, document, diagnosticCollection);
        } else {
            console.error('Error running substrate diff:', error);
            // Optionally, we could show an error message to the user if the CLI fails for unexpected reasons.
        }
    } finally {
        if (tempFilePath && fs.existsSync(tempFilePath)) {
            fs.unlinkSync(tempFilePath);
        }
    }
}

export function processOutput(output: string, document: vscode.TextDocument, diagnosticCollection: vscode.DiagnosticCollection) {
    try {
        const result = JSON.parse(output);
        const diagnostics: vscode.Diagnostic[] = [];

        // Parse breaking changes
        if (result.breaking_changes && Array.isArray(result.breaking_changes)) {
            for (const change of result.breaking_changes) {
                // Find approximate line number by searching for the field name extracted from the path
                const fieldName = extractFieldNameFromPath(change.path || '');
                const range = findRangeForField(document, fieldName);

                const message = `[BREAKING] ${change.rule_id || 'Unknown rule'}\n${change.message || 'Breaking change detected'}`;
                const diagnostic = new vscode.Diagnostic(range, message, vscode.DiagnosticSeverity.Error);
                diagnostic.source = 'Substrate';
                diagnostic.code = change.rule_id;

                diagnostics.push(diagnostic);
            }
        }

        // Can add warnings too if we want, but spec says Error (for BREAKING)
        if (result.warnings && Array.isArray(result.warnings)) {
            for (const warning of result.warnings) {
                const fieldName = extractFieldNameFromPath(warning.path || '');
                const range = findRangeForField(document, fieldName);

                const message = `[WARNING] ${warning.rule_id || 'Unknown rule'}\n${warning.message || 'Warning detected'}`;
                const diagnostic = new vscode.Diagnostic(range, message, vscode.DiagnosticSeverity.Warning);
                diagnostic.source = 'Substrate';
                diagnostic.code = warning.rule_id;

                diagnostics.push(diagnostic);
            }
        }

        diagnosticCollection.set(document.uri, diagnostics);

    } catch (e) {
        console.error('Failed to parse substrate JSON output:', e, output);
    }
}

export function extractFieldNameFromPath(jsonPath: string): string {
    // Example path: "components.schemas.User.properties.user_id"
    // Return "user_id"
    if (!jsonPath) return '';
    const parts = jsonPath.split('.');
    return parts[parts.length - 1];
}

export function findRangeForField(document: vscode.TextDocument, fieldName: string): vscode.Range {
    if (!fieldName) {
        return new vscode.Range(0, 0, 0, 0);
    }

    const text = document.getText();
    // A simple regex to find the field name (as a key in JSON/YAML, or a column in SQL)
    // This is a naive approach; a real parser would be better, but suffices for MVP.
    const lines = text.split('\n');

    for (let i = 0; i < lines.length; i++) {
        const line = lines[i];
        const index = line.indexOf(fieldName);
        if (index !== -1) {
            // Check if it's somewhat isolated (e.g., word boundary)
            const regex = new RegExp(`\\b${fieldName}\\b`);
            const match = regex.exec(line);
            if (match) {
                return new vscode.Range(i, match.index, i, match.index + fieldName.length);
            }
            return new vscode.Range(i, index, i, index + fieldName.length);
        }
    }

    return new vscode.Range(0, 0, 0, 0);
}
