import { EvalCase, Scorer } from '../types.js';
import { getAdapter } from '../../models/index.js';

interface JudgeResult {
  completeness: number;
  concreteness: number;
  risk_awareness: number;
  implementability: number;
  overall: number;
  reasoning: string;
}

export const llmJudgeScorer = (judgeModel: string = 'claude'): Scorer => {
  return {
    name: 'LLM Judge',
    max: 10,
    async score(text: string, _evalCase: EvalCase): Promise<number> {
      const adapter = getAdapter(judgeModel);

      const prompt = `You are evaluating a technical plan. Score it on each dimension from 0-10. Return ONLY valid JSON, no markdown, no code fences.

Plan:
${text}

Return: {"completeness": N, "concreteness": N, "risk_awareness": N, "implementability": N, "overall": N, "reasoning": "one sentence"}`;

      try {
        const response = await adapter.plan(prompt, 60000);
        const json = JSON.parse(response) as JudgeResult;
        return json.overall;
      } catch {
        return 0;
      }
    },
  };
};
