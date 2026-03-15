import { EvalCase, EvalReport, EvalScore, Scorer } from './types.js';
export declare function runScorersScore(text: string, evalCase: EvalCase, scorers: Scorer[]): Promise<EvalScore[]>;
export declare function computeOverallScore(scores: EvalScore[]): number;
export declare function buildReport(model: string, task: string, evalCase: EvalCase, scores: EvalScore[]): EvalReport;
//# sourceMappingURL=runner.d.ts.map