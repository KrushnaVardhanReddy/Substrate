import * as vscode from 'vscode';

export class SubstrateCodeActionProvider implements vscode.CodeActionProvider {
    public static readonly providedCodeActionKinds = [
        vscode.CodeActionKind.QuickFix
    ];

    public provideCodeActions(document: vscode.TextDocument, range: vscode.Range, context: vscode.CodeActionContext, token: vscode.CancellationToken): vscode.CodeAction[] {
        // Find if there's any substrate diagnostic in the current range
        const substrateDiagnostics = context.diagnostics.filter(diagnostic => diagnostic.source === 'Substrate');
        if (substrateDiagnostics.length === 0) {
            return [];
        }

        const actions: vscode.CodeAction[] = [];

        for (const diagnostic of substrateDiagnostics) {
            const fix = this.createFix(document, diagnostic.range);
            actions.push(fix);
        }

        return actions;
    }

    private createFix(document: vscode.TextDocument, range: vscode.Range): vscode.CodeAction {
        const fix = new vscode.CodeAction('AI Autofix: Add @deprecated (Mock)', vscode.CodeActionKind.QuickFix);
        fix.edit = new vscode.WorkspaceEdit();

        // Just add a generic deprecated comment above the line
        let commentPrefix = '# ';
        if (document.languageId === 'json') {
            commentPrefix = '// ';
        } else if (document.languageId === 'sql') {
            commentPrefix = '-- ';
        }

        // For GraphQL it's also #

        const indentation = document.lineAt(range.start.line).text.match(/^\s*/)?.[0] || '';
        const editRange = new vscode.Range(range.start.line, 0, range.start.line, 0);

        fix.edit.insert(document.uri, editRange.start, `${indentation}${commentPrefix}@deprecated\n`);

        return fix;
    }
}
