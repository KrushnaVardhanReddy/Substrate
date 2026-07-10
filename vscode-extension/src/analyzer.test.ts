import { describe, it, expect, vi } from 'vitest';
import { extractFieldNameFromPath, findRangeForField, processOutput, analyzeDocument } from './analyzer';
import * as cp from 'child_process';
import * as util from 'util';
import * as fs from 'fs';

vi.mock('child_process', () => ({
    execFile: vi.fn()
}));

vi.mock('util', () => ({
    promisify: (fn: any) => fn
}));

vi.mock('fs', () => ({
    writeFileSync: vi.fn(),
    existsSync: vi.fn(() => true),
    unlinkSync: vi.fn()
}));

describe('Analyzer', () => {
    describe('extractFieldNameFromPath', () => {
        it('should extract the field name from a simple path', () => {
            const field = extractFieldNameFromPath('user_id');
            expect(field).toBe('user_id');
        });

        it('should extract the field name from a nested path', () => {
            const field = extractFieldNameFromPath('components.schemas.User.properties.user_id');
            expect(field).toBe('user_id');
        });

        it('should return empty string for empty path', () => {
            const field = extractFieldNameFromPath('');
            expect(field).toBe('');
        });
    });

    describe('findRangeForField', () => {
        it('should find range for an exact match', () => {
            const document = {
                getText: () => 'type User {\n  user_id: ID!\n}'
            } as any;
            const range = findRangeForField(document, 'user_id');
            expect(range.startLine).toBe(1);
            expect(range.startChar).toBe(2);
            expect(range.endLine).toBe(1);
            expect(range.endChar).toBe(9);
        });

        it('should return 0,0,0,0 for no match', () => {
            const document = {
                getText: () => 'type User {\n  name: String\n}'
            } as any;
            const range = findRangeForField(document, 'user_id');
            expect(range.startLine).toBe(0);
        });

        it('should return 0,0,0,0 if field is empty', () => {
            const document = {
                getText: () => 'type User {\n  name: String\n}'
            } as any;
            const range = findRangeForField(document, '');
            expect(range.startLine).toBe(0);
        });
    });

    describe('processOutput', () => {
        it('should process JSON and set diagnostics', () => {
            const mockCollection = {
                set: vi.fn()
            } as any;
            const mockDoc = {
                uri: { fsPath: 'test.yaml', scheme: 'file' },
                getText: () => 'user_id: string'
            } as any;

            const output = JSON.stringify({
                breaking_changes: [
                    { path: 'user_id', rule_id: 'no-delete', message: 'deleted' }
                ],
                warnings: [
                    { path: 'user_id', rule_id: 'warn-1', message: 'warning' }
                ]
            });

            processOutput(output, mockDoc, mockCollection);
            expect(mockCollection.set).toHaveBeenCalledTimes(1);
            const args = mockCollection.set.mock.calls[0];
            expect(args[1].length).toBe(2); // 1 breaking, 1 warning
        });

        it('should handle invalid JSON output', () => {
            const mockCollection = { set: vi.fn() } as any;
            const mockDoc = { uri: { fsPath: 'test.yaml', scheme: 'file' }, getText: () => '' } as any;
            const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

            processOutput('invalid json', mockDoc, mockCollection);

            expect(consoleSpy).toHaveBeenCalled();
            consoleSpy.mockRestore();
        });
    });

    describe('analyzeDocument', () => {
        it('should return early if not a file scheme', async () => {
            const document = {
                uri: { scheme: 'untitled' }
            } as any;
            const collection = {} as any;
            await analyzeDocument(document, collection);
            expect(cp.execFile).not.toHaveBeenCalled();
        });

        it('should run child_process correctly for a file', async () => {
            const mockDoc = {
                uri: { scheme: 'file', fsPath: '/path/to/test.yaml' },
                getText: () => ''
            } as any;
            const mockCollection = { set: vi.fn() } as any;

            // util.promisify mock makes execFile return a Promise resolving to its first parameter structure
            vi.mocked(cp.execFile)
                .mockResolvedValueOnce({ stdout: 'git output' } as any)
                .mockResolvedValueOnce({ stdout: '{}' } as any);

            await analyzeDocument(mockDoc, mockCollection);

            expect(cp.execFile).toHaveBeenCalledTimes(2);
            expect(fs.writeFileSync).toHaveBeenCalled();
            expect(fs.unlinkSync).toHaveBeenCalled();
        });

        it('should handle substrate error with stdout (non-zero exit code)', async () => {
            const mockDoc = {
                uri: { scheme: 'file', fsPath: '/path/to/test.yaml' },
                getText: () => ''
            } as any;
            const mockCollection = { set: vi.fn() } as any;

            // util.promisify mock
            vi.mocked(cp.execFile).mockResolvedValueOnce({ stdout: 'git output' } as any);

            const error = new Error('Command failed') as any;
            error.stdout = JSON.stringify({ breaking_changes: [] });
            vi.mocked(cp.execFile).mockRejectedValueOnce(error);

            await analyzeDocument(mockDoc, mockCollection);

            expect(mockCollection.set).toHaveBeenCalledTimes(1);
        });

        it('should handle general execution error without stdout', async () => {
            const mockDoc = {
                uri: { scheme: 'file', fsPath: '/path/to/test.yaml' },
                getText: () => ''
            } as any;
            const mockCollection = { set: vi.fn() } as any;

            const error = new Error('Command failed');
            vi.mocked(cp.execFile).mockRejectedValueOnce(error);

            const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

            await analyzeDocument(mockDoc, mockCollection);

            expect(consoleSpy).toHaveBeenCalledWith('Error running substrate diff:', error);
            consoleSpy.mockRestore();
        });
    });
});
