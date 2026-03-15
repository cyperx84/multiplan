import { ModelAdapter } from './types.js';
export declare class GeminiAdapter implements ModelAdapter {
    id: string;
    name: string;
    available(): Promise<boolean>;
    plan(prompt: string, timeoutMs?: number): Promise<string>;
    private commandExists;
    private runCommand;
}
//# sourceMappingURL=gemini.d.ts.map