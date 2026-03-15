export interface EvalCase {
  task: string;
  requirements?: string;
  constraints?: string;
  expectedTopics?: string[];
  minScore?: number;
}

export interface EvalScore {
  name: string;
  score: number;
  max: number;
  pass: boolean;
  details?: string;
}

export interface EvalReport {
  task: string;
  model: string;
  scores: EvalScore[];
  overallScore: number;
  pass: boolean;
  summary: string;
  markdown: string;
}

export interface Scorer {
  name: string;
  max: number;
  score(text: string, evalCase: EvalCase): Promise<number>;
}
