import { ClaudeAdapter } from './claude.js';
import { GeminiAdapter } from './gemini.js';
import { CodexAdapter } from './codex.js';
import { GLMAdapter } from './glm.js';
export { ClaudeAdapter } from './claude.js';
export { GeminiAdapter } from './gemini.js';
export { CodexAdapter } from './codex.js';
export { GLMAdapter } from './glm.js';
const adapters = {
    claude: new ClaudeAdapter(),
    gemini: new GeminiAdapter(),
    codex: new CodexAdapter(),
    glm5: new GLMAdapter(),
};
export async function getAvailableModels() {
    const available = [];
    for (const [id, adapter] of Object.entries(adapters)) {
        if (await adapter.available()) {
            available.push(id);
        }
    }
    return available;
}
export function getAdapter(id) {
    const adapter = adapters[id];
    if (!adapter) {
        throw new Error(`Unknown model: ${id}`);
    }
    return adapter;
}
export function getAllAdapters() {
    return adapters;
}
//# sourceMappingURL=index.js.map