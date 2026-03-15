import { EvalCase, EvalReport } from './types.js';
export declare function evalPlan(planText: string, evalCase: EvalCase, opts?: {
    judge?: string;
}): Promise<EvalReport>;
export declare function evalRun(runDir: string, evalCase: EvalCase, opts?: {
    judge?: string;
}): Promise<Record<string, EvalReport>>;
export { EvalReport, EvalCase } from './types.js';
//# sourceMappingURL=index.d.ts.map