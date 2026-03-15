import { ClaudeAdapter } from './claude.js';
import { GeminiAdapter } from './gemini.js';
import { CodexAdapter } from './codex.js';
import { GLMAdapter } from './glm.js';
import { ModelAdapter } from './types.js';

export { ModelAdapter, ModelResult } from './types.js';
export { ClaudeAdapter } from './claude.js';
export { GeminiAdapter } from './gemini.js';
export { CodexAdapter } from './codex.js';
export { GLMAdapter } from './glm.js';

const adapters: Record<string, ModelAdapter> = {
  claude: new ClaudeAdapter(),
  gemini: new GeminiAdapter(),
  codex: new CodexAdapter(),
  glm5: new GLMAdapter(),
};

export async function getAvailableModels(): Promise<string[]> {
  const available: string[] = [];

  for (const [id, adapter] of Object.entries(adapters)) {
    if (await adapter.available()) {
      available.push(id);
    }
  }

  return available;
}

export function getAdapter(id: string): ModelAdapter {
  const adapter = adapters[id];
  if (!adapter) {
    throw new Error(`Unknown model: ${id}`);
  }
  return adapter;
}

export function getAllAdapters(): Record<string, ModelAdapter> {
  return adapters;
}
