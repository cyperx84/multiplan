import { EvalCase, Scorer } from '../types.js';

const REQUIRED_SECTIONS = [
  'Overview',
  'Architecture',
  'Implementation',
  'Trade-off',
  'Risk',
];

export const coverageScorer: Scorer = {
  name: 'Coverage',
  max: 1,
  async score(text: string, _evalCase: EvalCase): Promise<number> {
    let found = 0;
    for (const section of REQUIRED_SECTIONS) {
      if (new RegExp(`##\\s+${section}`, 'i').test(text)) {
        found++;
      }
    }
    return Math.min(found / REQUIRED_SECTIONS.length, 1);
  },
};
