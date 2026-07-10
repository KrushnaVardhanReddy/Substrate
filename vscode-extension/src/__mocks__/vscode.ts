export const workspace = {
  getWorkspaceFolder: () => null,
};
export const Diagnostic = class {};
export const DiagnosticSeverity = {
  Error: 0,
  Warning: 1,
  Information: 2,
  Hint: 3
};
export const Range = class {
  constructor(public startLine: number, public startChar: number, public endLine: number, public endChar: number) {}
};
export const Uri = {
  file: (path: string) => ({ fsPath: path, scheme: 'file' })
};
export const Position = class {
  constructor(public line: number, public character: number) {}
};
export const CodeActionKind = {
  QuickFix: 'QuickFix'
};
export const CodeAction = class {
  edit: any;
  constructor(public title: string, public kind: string) {}
};
export const WorkspaceEdit = class {
  insert = (uri: any, position: any, text: string) => {};
};
