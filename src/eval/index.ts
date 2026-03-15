import { promises as fs } from 'fs';
import { join, dirname, basename } from 'path';
import { EvalCase, EvalReport } from './types.js';
import { runScorersScore, buildReport } from './runner.js';
import { coverageScorer } from './scorers/coverage.js';
import { specificitySc } from './scorers/specificity.js';
import { actionableScorer } from './scorers/actionable.js';
import { llmJudgeScorer } from './scorers/llm-judge.js';

const DEFAULT_SCORERS = [
  coverageScorer,
  specificitySc,
  actionableScorer,
];

export async function evalPlan(
  planText: string,
  evalCase: EvalCase,
  opts?: { judge?: string }
): Promise<EvalReport> {
  const scorers = [...DEFAULT_SCORERS];

  if (opts?.judge) {
    scorers.push(llmJudgeScorer(opts.judge));
  }

  const scores = await runScorersScore(planText, evalCase, scorers);
  const report = buildReport('plan', evalCase.task, evalCase, scores);

  return report;
}

export async function evalRun(
  runDir: string,
  evalCase: EvalCase
): Promise<Record<string, EvalReport>> {
  const reports: Record<string, EvalReport> = {};

  // Eval all plan-*.md files
  try {
    const files = await fs.readdir(runDir);
    for (const file of files) {
      if (file.startsWith('plan-') && file.endsWith('.md')) {
        const modelId = file.replace('plan-', '').replace('.md', '');
        const filePath = join(runDir, file);
        const content = await fs.readFile(filePath, 'utf-8');

        const report = await evalPlan(content, evalCase);
        report.model = modelId;
        reports[modelId] = report;
      }
    }

    // Eval final-plan.md if it exists
    try {
      const finalPath = join(runDir, 'final-plan.md');
      const finalContent = await fs.readFile(finalPath, 'utf-8');
      const report = await evalPlan(finalContent, evalCase);
      report.model = 'final';
      reports.final = report;
    } catch {
      // final plan doesn't exist, skip
    }
  } catch {
    // directory doesn't exist
  }

  return reports;
}

export { EvalReport, EvalCase } from './types.js';
