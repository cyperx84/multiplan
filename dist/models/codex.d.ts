import { ModelAdapter } from './types.js';
export declare class CodexAdapter implements ModelAdapter {
    id: string;
    name: string;
    available(): Promise<boolean>;
    plan(prompt: string, timeoutMs?: number): Promise<string>;
    private commandExists;
    private runCommand;
}
//# sourceMappingURL=codex.d.ts.map