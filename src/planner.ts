import { promises as fs } from 'fs';
import { homedir } from 'os';
import { join } from 'path';
import { getAdapter, getAvailableModels, ModelResult } from './models/index.js';
import { loadAndRender } from './prompts/loader.js';

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

function generateRunId(): string {
  const date = new Date();
  const timestamp = date.toISOString().replace(/[:.]/g, '').split('Z')[0];
  const random = Math.random().toString(36).substring(2, 8);
  return `${timestamp}-${random}`;
}

async function ensureDir(path: string): Promise<void> {
  try {
    await fs.mkdir(path, { recursive: true });
  } catch {
    // ignore
  }
}

async function writeFile(path: string, content: string): Promise<void> {
  await fs.writeFile(path, content, 'utf-8');
}

async function readFile(path: string): Promise<string> {
  return fs.readFile(path, 'utf-8');
}

export async function run(config: PlannerConfig): Promise<PlannerRun> {
  const runId = generateRunId();
  const requirements = config.requirements || 'None specified.';
  const constraints = config.constraints || 'None specified.';
  const timeoutMs = config.timeoutMs || 120000;

  // Determine output directory
  const outputDir =
    config.outputDir ||
    join(homedir(), '.multiplan', 'runs', runId);
  await ensureDir(outputDir);

  if (config.verbose) {
    console.log(`[multiplan] Run ID: ${runId}`);
    console.log(`[multiplan] Output: ${outputDir}`);
  }

  // Phase 1: Parallel planning
  if (config.verbose) {
    console.log('[multiplan] Phase 1 — Running models in parallel...');
  }

  const planPrompt = loadAndRender('plan', {
    TASK: config.task,
    REQUIREMENTS: requirements,
    CONSTRAINTS: constraints,
  });

  const availableModels = await getAvailableModels();
  const modelIds = config.models || availableModels;
  const results: ModelResult[] = [];

  // Run all models in parallel
  const planPromises = modelIds.map(async (modelId) => {
    const adapter = getAdapter(modelId);
    const startTime = Date.now();

    try {
      const plan = await adapter.plan(planPrompt, timeoutMs);
      const durationMs = Date.now() - startTime;
      const result: ModelResult = {
        modelId,
        modelName: adapter.name,
        plan,
        durationMs,
      };
      results.push(result);

      // Write plan to disk
      const filename = join(outputDir, `plan-${modelId}.md`);
      await writeFile(filename, plan);

      if (config.verbose) {
        console.log(
          `[multiplan]   ✓ ${adapter.name} done (${durationMs}ms)`
        );
      }
    } catch (error) {
      const durationMs = Date.now() - startTime;
      const result: ModelResult = {
        modelId,
        modelName: adapter.name,
        plan: `[Error: ${error instanceof Error ? error.message : String(error)}]`,
        durationMs,
        error: error instanceof Error ? error.message : String(error),
      };
      results.push(result);

      if (config.verbose) {
        console.log(
          `[multiplan]   ✗ ${adapter.name} failed: ${result.error}`
        );
      }
    }
  });

  await Promise.all(planPromises);

  if (config.verbose) {
    console.log('[multiplan] Phase 1 complete');
  }

  // Phase 2: Cross-examination (debate)
  if (config.verbose) {
    console.log('[multiplan] Phase 2 — Cross-examination...');
  }

  const debateModel = config.debateModel || 'claude';
  const debateAdapter = getAdapter(debateModel);

  const planContents: Record<string, string> = {};
  for (let i = 0; i < results.length; i++) {
    const letter = String.fromCharCode(65 + i); // A, B, C, D
    planContents[`PLAN_${letter}`] = results[i].plan;
  }

  const debatePrompt = loadAndRender('debate', {
    TASK: config.task,
    ...planContents,
  });

  let debate: string;
  try {
    debate = await debateAdapter.plan(debatePrompt, timeoutMs);
    if (config.verbose) {
      console.log(`[multiplan]   ✓ Debate complete (via ${debateAdapter.name})`);
    }
  } catch (error) {
    debate = `[Debate failed: ${error instanceof Error ? error.message : String(error)}]`;
    if (config.verbose) {
      console.log(`[multiplan]   ✗ Debate failed: ${debate}`);
    }
  }

  await writeFile(join(outputDir, 'debate.md'), debate);

  // Phase 3: Convergence
  if (config.verbose) {
    console.log('[multiplan] Phase 3 — Convergence...');
  }

  const convergeModel = config.convergeModel || 'claude';
  const convergeAdapter = getAdapter(convergeModel);

  const convergePrompt = loadAndRender('converge', {
    TASK: config.task,
    ...planContents,
    DEBATE: debate,
  });

  let finalPlan: string;
  try {
    finalPlan = await convergeAdapter.plan(convergePrompt, timeoutMs);
    if (config.verbose) {
      console.log(`[multiplan]   ✓ Convergence complete (via ${convergeAdapter.name})`);
    }
  } catch (error) {
    finalPlan = `[Convergence failed: ${error instanceof Error ? error.message : String(error)}]`;
    if (config.verbose) {
      console.log(`[multiplan]   ✗ Convergence failed: ${finalPlan}`);
    }
  }

  // Add metadata header
  const header = `# Multimodel Plan: ${config.task}\n\n> Generated: ${new Date().toISOString()}\n> Models: ${results.map((r) => r.modelName).join(', ')}\n\n---\n\n`;
  const fullPlan = header + finalPlan;

  await writeFile(join(outputDir, 'final-plan.md'), fullPlan);

  if (config.verbose) {
    console.log('[multiplan] Phase 3 complete');
  }

  return {
    runId,
    outputDir,
    plans: results,
    debate,
    finalPlan: fullPlan,
  };
}
