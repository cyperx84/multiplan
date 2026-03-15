import { EvalCase, EvalReport, EvalScore, Scorer } from './types.js';

export async function runScorersScore(
  text: string,
  evalCase: EvalCase,
  scorers: Scorer[]
): Promise<EvalScore[]> {
  const scores: EvalScore[] = [];

  for (const scorer of scorers) {
    const score = await scorer.score(text, evalCase);
    const normalized = score / scorer.max;
    const pass = normalized >= 0.5;

    scores.push({
      name: scorer.name,
      score: normalized,
      max: 1,
      pass,
    });
  }

  return scores;
}

export function computeOverallScore(scores: EvalScore[]): number {
  if (scores.length === 0) return 0;
  const sum = scores.reduce((acc, s) => acc + s.score, 0);
  return sum / scores.length;
}

export function buildReport(
  model: string,
  task: string,
  evalCase: EvalCase,
  scores: EvalScore[]
): EvalReport {
  const overallScore = computeOverallScore(scores);
  const minScore = evalCase.minScore || 6;
  const pass = overallScore * 10 >= minScore;

  const summary = `Model: ${model} | Overall: ${(overallScore * 10).toFixed(1)}/10 | ${pass ? 'PASS' : 'FAIL'}`;

  let markdown = `# Evaluation Report\n\n`;
  markdown += `**Task:** ${task}\n`;
  markdown += `**Model:** ${model}\n`;
  markdown += `**Overall Score:** ${(overallScore * 10).toFixed(1)}/10 ${pass ? '✓' : '✗'}\n\n`;

  markdown += `## Scores\n\n`;
  for (const score of scores) {
    const icon = score.pass ? '✓' : '✗';
    markdown += `- **${score.name}:** ${(score.score * 10).toFixed(1)}/10 ${icon}\n`;
    if (score.details) {
      markdown += `  - ${score.details}\n`;
    }
  }

  return {
    task,
    model,
    scores,
    overallScore,
    pass,
    summary,
    markdown,
  };
}
