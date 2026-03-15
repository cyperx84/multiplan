# Contributing to multiplan

## Development Setup

```bash
git clone https://github.com/cyperx84/multiplan.git
cd multiplan
npm install
npm run build
```

## Building

```bash
# TypeScript → JavaScript
npm run build

# Watch mode
npm run dev

# Tests
npm test
```

## Project Structure

```
src/
├── cli.ts                  # Command-line interface (commander)
├── planner.ts              # Core orchestration (3 phases)
├── models/                 # Model adapters
│   ├── types.ts           # ModelAdapter interface
│   ├── claude.ts          # Claude adapter
│   ├── gemini.ts          # Gemini adapter
│   ├── codex.ts           # Codex adapter
│   ├── glm.ts             # GLM-5 adapter
│   └── index.ts           # Model registry
├── prompts/               # Prompt templates
│   ├── plan.md            # Planning prompt
│   ├── debate.md          # Cross-examination prompt
│   ├── converge.md        # Synthesis prompt
│   └── loader.ts          # Template rendering
└── eval/                  # Evaluation framework
    ├── types.ts           # EvalCase, EvalReport types
    ├── runner.ts          # Scorer runner + report builder
    ├── index.ts           # evalPlan() and evalRun()
    └── scorers/
        ├── coverage.ts    # Section coverage scorer
        ├── specificity.ts # Concrete vs vague language scorer
        ├── actionable.ts  # Numbered steps, code blocks scorer
        └── llm-judge.ts   # LLM-based grading scorer
```

## Adding a New Model

1. Create `src/models/new-model.ts`:

```typescript
import { ModelAdapter } from './types.js';

export class NewModelAdapter implements ModelAdapter {
  id = 'newmodel';
  name = 'New Model';

  async available(): Promise<boolean> {
    // Check if the model is available
    return true;
  }

  async plan(prompt: string, timeoutMs?: number): Promise<string> {
    // Implement planning logic
    // Call external API or CLI, pipe prompt, return response
  }
}
```

2. Register in `src/models/index.ts`:

```typescript
import { NewModelAdapter } from './new-model.js';

const adapters: Record<string, ModelAdapter> = {
  // ...
  newmodel: new NewModelAdapter(),
};
```

## Adding a New Scorer

1. Create `src/eval/scorers/new-scorer.ts`:

```typescript
import { EvalCase, Scorer } from '../types.js';

export const newScorer: Scorer = {
  name: 'New Scorer',
  max: 1,  // or 10 for LLM scorers
  async score(text: string, evalCase: EvalCase): Promise<number> {
    // Implement scoring logic
    // Return score between 0 and max
  },
};
```

2. Add to defaults in `src/eval/index.ts`:

```typescript
import { newScorer } from './scorers/new-scorer.js';

const DEFAULT_SCORERS = [
  // ...
  newScorer,
];
```

## Testing

Tests use Node's built-in test runner. Structural scorers have unit tests:

```bash
npm test
```

Current test coverage:
- Coverage scorer (required sections)
- Specificity scorer (concrete vs vague ratio)
- Actionable scorer (numbered items, code blocks)

## Submitting Changes

1. Create a feature branch: `git checkout -b feature/my-feature`
2. Commit with clear messages: `git commit -m "feat: add new model adapter"`
3. Push and create a PR
4. Ensure `npm run build` and `npm test` pass

## Release Process

1. Update version in `package.json`
2. Create changelog entry
3. Tag release: `git tag v0.x.0`
4. Push tags: `git push origin --tags`
5. Publish to npm: `npm publish`

## Questions?

Open an issue or discussion on GitHub. All contributions welcome!
