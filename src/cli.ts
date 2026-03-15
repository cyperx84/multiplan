#!/usr/bin/env node

import { Command } from 'commander';
import { run } from './planner.js';
import { evalRun, evalPlan } from './eval/index.js';
import { promises as fs } from 'fs';
import { basename } from 'path';

const program = new Command();

program
  .name('multiplan')
  .description('4-model parallel planning workflow with eval framework')
  .version('0.1.0');

// multiplan <task> [options]
program
  .command('plan <task>')
  .description('Run multi-model planning workflow')
  .option('--req <requirements>', 'Requirements')
  .option('--con <constraints>', 'Constraints')
  .option('--out <dir>', 'Output directory')
  .option('--models <list>', 'Comma-separated models (claude,gemini,codex,glm5)')
  .option('--debate-model <model>', 'Model for debate phase')
  .option('--converge-model <model>', 'Model for convergence phase')
  .option('--timeout <ms>', 'Per-model timeout in milliseconds')
  .option('--verbose', 'Verbose output')
  .action(async (task, options) => {
    try {
      const config = {
        task,
        requirements: options.req,
        constraints: options.con,
        outputDir: options.out,
        models: options.models ? options.models.split(',') : undefined,
        debateModel: options.debateModel,
        convergeModel: options.convergeModel,
        timeoutMs: options.timeout ? parseInt(options.timeout) : undefined,
        verbose: options.verbose,
      };

      const result = await run(config);

      console.log('\n════════════════════════════════════════');
      console.log(' Multimodel Planning Complete');
      console.log('════════════════════════════════════════\n');
      console.log(` Task: ${task}\n`);
      console.log(` Outputs:`);
      for (const plan of result.plans) {
        console.log(`   ${plan.modelName}: ${result.outputDir}/plan-${plan.modelId}.md`);
      }
      console.log(`   Debate:     ${result.outputDir}/debate.md`);
      console.log(`   Final Plan: ${result.outputDir}/final-plan.md\n`);
      console.log(` → Final plan saved to: ${result.outputDir}/final-plan.md\n`);
      console.log(result.finalPlan);
    } catch (error) {
      console.error('Error:', error);
      process.exit(1);
    }
  });

// multiplan eval <file-or-dir> [options]
program
  .command('eval <pathOrDir>')
  .description('Evaluate a plan or run directory')
  .option('--fixture <path>', 'Fixture JSON file')
  .option('--judge <model>', 'Model for LLM judge (claude, gemini, etc.)')
  .option('--json', 'Output as JSON')
  .action(async (pathOrDir, options) => {
    try {
      // Load fixture
      let evalCase: any = {
        task: 'Unknown task',
      };

      if (options.fixture) {
        const fixtureContent = await fs.readFile(options.fixture, 'utf-8');
        evalCase = JSON.parse(fixtureContent);
      }

      // Check if it's a directory or file
      let stat;
      try {
        stat = await fs.stat(pathOrDir);
      } catch {
        console.error(`Error: Path not found: ${pathOrDir}`);
        process.exit(1);
      }

      if (stat.isDirectory()) {
        const reports = await evalRun(pathOrDir, evalCase);

        if (options.json) {
          console.log(JSON.stringify(reports, null, 2));
        } else {
          for (const [model, report] of Object.entries(reports)) {
            console.log(report.markdown);
            console.log('---\n');
          }
        }
      } else {
        const planContent = await fs.readFile(pathOrDir, 'utf-8');
        const report = await evalPlan(planContent, evalCase, { judge: options.judge });

        if (options.json) {
          console.log(JSON.stringify(report, null, 2));
        } else {
          console.log(report.markdown);
        }
      }
    } catch (error) {
      console.error('Error:', error);
      process.exit(1);
    }
  });

// multiplan skill (stub for now)
program
  .command('skill')
  .description('Generate OpenClaw SKILL.md')
  .action(() => {
    console.log('Skill generation not yet implemented.');
  });

// multiplan integrations (stub for now)
program
  .command('integrations')
  .description('Print integration snippets')
  .action(() => {
    console.log('Integration generation not yet implemented.');
  });

program.parse(process.argv);
