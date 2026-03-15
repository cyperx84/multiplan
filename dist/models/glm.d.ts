import { ModelAdapter } from './types.js';
export declare class GLMAdapter implements ModelAdapter {
    id: string;
    name: string;
    private apiUrl;
    available(): Promise<boolean>;
    plan(prompt: string, timeoutMs?: number): Promise<string>;
    private getApiKey;
}
//# sourceMappingURL=glm.d.ts.map