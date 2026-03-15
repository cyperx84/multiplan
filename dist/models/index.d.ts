import { ModelAdapter } from './types.js';
export { ModelAdapter, ModelResult } from './types.js';
export { ClaudeAdapter } from './claude.js';
export { GeminiAdapter } from './gemini.js';
export { CodexAdapter } from './codex.js';
export { GLMAdapter } from './glm.js';
export declare function getAvailableModels(): Promise<string[]>;
export declare function getAdapter(id: string): ModelAdapter;
export declare function getAllAdapters(): Record<string, ModelAdapter>;
//# sourceMappingURL=index.d.ts.map