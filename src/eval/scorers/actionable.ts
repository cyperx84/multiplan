import { EvalCase, Scorer } from '../types.js';

export const actionableScorer: Scorer = {
  name: 'Actionable',
  max: 1,
  async score(text: string, _evalCase: EvalCase): Promise<number> {
    const hasNumberedList = /^\d+\./m.test(text);
    const hasCodeFences = /```/g.test(text);
    const hasBulletPoints = /^[-*]/m.test(text);

    const codeBlockCount = (text.match(/```/g) || []).length / 2;
    const numberedItems = (text.match(/^\d+\./gm) || []).length;

    let score = 0;

    if (hasNumberedList) score += 0.3;
    if (hasCodeFences) score += 0.3;
    if (hasBulletPoints) score += 0.2;

    // Bonus for density
    if (codeBlockCount >= 2) score += 0.1;
    if (numberedItems >= 5) score += 0.1;

    return Math.min(score, 1);
  },
};
