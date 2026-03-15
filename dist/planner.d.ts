import { ModelResult } from './models/index.js';
export interface PlannerConfig {
    task: string;
    requirements?: string;
    constraints?: string;
    models?: string[];
    outputDir?: string;
    debateModel?: string;
    convergeModel?: string;
    timeoutMs?: number;
    verbose?: boolean;
}
export interface PlannerRun {
    runId: string;
    outputDir: string;
    plans: ModelResult[];
    debate: string;
    finalPlan: string;
}
export declare function run(config: PlannerConfig): Promise<PlannerRun>;
//# sourceMappingURL=planner.d.ts.map