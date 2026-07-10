import * as vscode from 'vscode';
import { analyzeDocument } from './analyzer';
import { SubstrateCodeActionProvider } from './codeActions';

let diagnosticCollection: vscode.DiagnosticCollection;

export function activate(context: vscode.ExtensionContext) {
    console.log('Substrate VSCode extension is now active!');

    // Register diagnostic collection
    diagnosticCollection = vscode.languages.createDiagnosticCollection('substrate');
    context.subscriptions.push(diagnosticCollection);

    // Register code action provider
    const selector: vscode.DocumentSelector = [
        { language: 'yaml' },
        { language: 'json' },
        { language: 'graphql' },
        { language: 'sql' }
    ];

    context.subscriptions.push(
        vscode.languages.registerCodeActionsProvider(selector, new SubstrateCodeActionProvider(), {
            providedCodeActionKinds: SubstrateCodeActionProvider.providedCodeActionKinds
        })
    );

    // Analyze on save
    context.subscriptions.push(
        vscode.workspace.onDidSaveTextDocument(async (document) => {
            if (isSupportedLanguage(document.languageId)) {
                await analyzeDocument(document, diagnosticCollection);
            }
        })
    );

    // Initial analysis if an editor is open
    if (vscode.window.activeTextEditor) {
        const document = vscode.window.activeTextEditor.document;
        if (isSupportedLanguage(document.languageId)) {
            analyzeDocument(document, diagnosticCollection);
        }
    }
}

function isSupportedLanguage(languageId: string): boolean {
    return ['yaml', 'json', 'graphql', 'sql'].includes(languageId);
}

export function deactivate() {
    if (diagnosticCollection) {
        diagnosticCollection.clear();
        diagnosticCollection.dispose();
    }
}
