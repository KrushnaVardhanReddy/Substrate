import { describe, it, expect, vi } from 'vitest';
import * as vscode from 'vscode';
import { activate, deactivate } from './extension';
import { SubstrateCodeActionProvider } from './codeActions';

// Update mock object to have all required functions and variables for testing
vi.mock('vscode', async () => {
    const actual = await vi.importActual<typeof import('./__mocks__/vscode')>('./__mocks__/vscode');
    return {
        ...actual,
        languages: {
            createDiagnosticCollection: vi.fn(() => ({
                clear: vi.fn(),
                dispose: vi.fn()
            })),
            registerCodeActionsProvider: vi.fn()
        },
        workspace: {
            onDidSaveTextDocument: vi.fn(),
            getWorkspaceFolder: () => null
        },
        window: {
            activeTextEditor: null
        }
    };
});

describe('Extension', () => {
    it('should activate properly', () => {
        const context = {
            subscriptions: {
                push: vi.fn()
            }
        } as any;

        activate(context);

        expect(vscode.languages.createDiagnosticCollection).toHaveBeenCalledWith('substrate');
        expect(vscode.languages.registerCodeActionsProvider).toHaveBeenCalled();
        expect(context.subscriptions.push).toHaveBeenCalledTimes(3);
    });

    it('should deactivate properly', () => {
        const context = {
            subscriptions: {
                push: vi.fn()
            }
        } as any;

        activate(context);

        // This will clear the module scope variable `diagnosticCollection`
        expect(() => deactivate()).not.toThrow();
    });
});
