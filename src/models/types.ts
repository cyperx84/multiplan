export interface ModelAdapter {
  id: string;
  name: string;
  available(): Promise<boolean>;
  plan(prompt: string, timeoutMs?: number): Promise<string>;
}

export interface ModelResult {
  modelId: string;
  modelName: string;
  plan: string;
  durationMs: number;
  error?: string;
}
