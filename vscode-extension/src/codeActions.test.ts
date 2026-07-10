import { describe, it, expect, vi } from 'vitest';
import { SubstrateCodeActionProvider } from './codeActions';

describe('CodeActionProvider', () => {
    it('should return quick fix for substrate diagnostics', () => {
        const provider = new SubstrateCodeActionProvider();
        const document = {
            languageId: 'yaml',
            uri: { fsPath: 'test.yaml' },
            lineAt: () => ({ text: '  user_id: string' })
        } as any;
        const range = { startLine: 0, start: { line: 0, character: 0 } } as any;
        const context = {
            diagnostics: [
                { source: 'Substrate', range }
            ]
        } as any;
        const token = {} as any;

        const actions = provider.provideCodeActions(document, range, context, token);
        expect(actions.length).toBe(1);
        expect(actions[0].title).toBe('AI Autofix: Add @deprecated (Mock)');
        expect(actions[0].edit).toBeDefined();
    });

    it('should ignore non-substrate diagnostics', () => {
        const provider = new SubstrateCodeActionProvider();
        const document = {} as any;
        const range = {} as any;
        const context = {
            diagnostics: [
                { source: 'eslint', range }
            ]
        } as any;
        const token = {} as any;

        const actions = provider.provideCodeActions(document, range, context, token);
        expect(actions.length).toBe(0);
    });
});
