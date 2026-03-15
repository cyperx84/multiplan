#!/usr/bin/env node
import { Command } from 'commander';
import { run } from './planner.js';
import { evalRun, evalPlan } from './eval/index.js';
import { writeSkill } from './integrations/openclaw.js';
import { generateClaudeCodeSnippet, generateCodexAgentSnippet } from './integrations/claude-code.js';
import { promises as fs } from 'fs';
import { resolve } from 'path';
const program = new Command();
program
    .name('multiplan')
    .description('4-model parallel planning workflow with eval framework')
    .version('0.1.0');
// ── Helper: run planning (shared between default + plan subcommand) ────────────
async function runPlanning(task, options) {
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
        if (!plan.error) {
            console.log(`   ${plan.modelName}: ${result.outputDir}/plan-${plan.modelId}.md`);
        }
        else {
            console.log(`   ${plan.modelName}: FAILED — ${plan.error}`);
        }
    }
    console.log(`   Debate:     ${result.outputDir}/debate.md`);
    console.log(`   Final Plan: ${result.outputDir}/final-plan.md\n`);
    console.log(result.finalPlan);
}
// ── multiplan <task> (default — no subcommand needed) ─────────────────────────
program
    .argument('[task]', 'Task to plan (runs planning workflow directly)')
    .option('--req <requirements>', 'Requirements')
    .option('--con <constraints>', 'Constraints')
    .option('--out <dir>', 'Output directory')
    .option('--models <list>', 'Comma-separated models (claude,gemini,codex,glm5)')
    .option('--debate-model <model>', 'Model for debate phase')
    .option('--converge-model <model>', 'Model for convergence phase')
    .option('--timeout <ms>', 'Per-model timeout in milliseconds')
    .option('--verbose', 'Verbose output')
    .action(async (task, options) => {
    // Only fire if a task was provided and no subcommand was matched
    if (task) {
        try {
            await runPlanning(task, options);
        }
        catch (error) {
            console.error('Error:', error);
            process.exit(1);
        }
    }
    // If no task, fall through to show help (handled by subcommands)
});
// ── multiplan plan <task> (explicit subcommand) ───────────────────────────────
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
        await runPlanning(task, options);
    }
    catch (error) {
        console.error('Error:', error);
        process.exit(1);
    }
});
// ── multiplan eval <file-or-dir> ──────────────────────────────────────────────
program
    .command('eval <pathOrDir>')
    .description('Evaluate a plan file or run directory')
    .option('--fixture <path>', 'Fixture JSON file (task, requirements, expectedTopics, minScore)')
    .option('--judge <model>', 'Model for LLM judge scorer (claude, gemini)')
    .option('--json', 'Output as JSON')
    .option('--no-judge', 'Skip LLM judge scorer (faster, structural only)')
    .action(async (pathOrDir, options) => {
    try {
        // Load fixture or build minimal evalCase
        let evalCase = { task: 'Unknown task' };
        if (options.fixture) {
            const fixtureContent = await fs.readFile(resolve(options.fixture), 'utf-8');
            evalCase = JSON.parse(fixtureContent);
        }
        let stat;
        try {
            stat = await fs.stat(pathOrDir);
        }
        catch {
            console.error(`Error: Path not found: ${pathOrDir}`);
            process.exit(1);
        }
        const evalOpts = {
            judge: options.judge !== false ? (options.judge ?? 'claude') : undefined,
        };
        if (stat.isDirectory()) {
            const reports = await evalRun(pathOrDir, evalCase, evalOpts);
            if (options.json) {
                console.log(JSON.stringify(reports, null, 2));
            }
            else {
                for (const report of Object.values(reports)) {
                    console.log(report.markdown);
                    console.log('---\n');
                }
                // Summary table
                console.log('## Summary\n');
                console.log('| Model | Score | Pass |');
                console.log('|-------|-------|------|');
                for (const [model, report] of Object.entries(reports)) {
                    const score = (report.overallScore * 10).toFixed(1);
                    const pass = report.pass ? '✅' : '❌';
                    console.log(`| ${model} | ${score}/10 | ${pass} |`);
                }
            }
        }
        else {
            const planContent = await fs.readFile(pathOrDir, 'utf-8');
            const report = await evalPlan(planContent, evalCase, evalOpts);
            if (options.json) {
                console.log(JSON.stringify(report, null, 2));
            }
            else {
                console.log(report.markdown);
            }
        }
    }
    catch (error) {
        console.error('Error:', error);
        process.exit(1);
    }
});
// ── multiplan skill ────────────────────────────────────────────────────────────
program
    .command('skill')
    .description('Generate/update the OpenClaw SKILL.md at ~/openclaw/skills/multiplan/')
    .action(() => {
    try {
        writeSkill();
    }
    catch (error) {
        console.error('Error writing skill:', error);
        process.exit(1);
    }
});
// ── multiplan integrations ─────────────────────────────────────────────────────
program
    .command('integrations')
    .description('Print integration snippets for Claude Code, Codex, and other agents')
    .option('--claude-code', 'Print CLAUDE.md snippet only')
    .option('--codex', 'Print .codex/agent.md snippet only')
    .action((options) => {
    if (options.claudeCode) {
        console.log(generateClaudeCodeSnippet());
    }
    else if (options.codex) {
        console.log(generateCodexAgentSnippet());
    }
    else {
        console.log('## CLAUDE.md snippet\n');
        console.log('```markdown');
        console.log(generateClaudeCodeSnippet());
        console.log('```\n');
        console.log('## .codex/agent.md snippet\n');
        console.log('```markdown');
        console.log(generateCodexAgentSnippet());
        console.log('```');
    }
});
program.parse(process.argv);
//# sourceMappingURL=cli.js.map