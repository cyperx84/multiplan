import { getAdapter } from '../../models/index.js';
export const llmJudgeScorer = (judgeModel = 'claude') => {
    return {
        name: 'LLM Judge',
        max: 10,
        async score(text, _evalCase) {
            const adapter = getAdapter(judgeModel);
            const prompt = `You are evaluating a technical plan. Score it on each dimension from 0-10. Return ONLY valid JSON, no markdown, no code fences.

Plan:
${text}

Return: {"completeness": N, "concreteness": N, "risk_awareness": N, "implementability": N, "overall": N, "reasoning": "one sentence"}`;
            try {
                const response = await adapter.plan(prompt, 60000);
                const json = JSON.parse(response);
                return json.overall;
            }
            catch {
                return 0;
            }
        },
    };
};
//# sourceMappingURL=llm-judge.js.map