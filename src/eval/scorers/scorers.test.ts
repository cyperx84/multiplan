/**
 * Unit tests for structural scorers — no model calls required.
 * Run with: node --test dist/eval/scorers/scorers.test.js
 */

import { strict as assert } from 'assert';
import { describe, it } from 'node:test';
import { coverageScorer } from './coverage.js';
import { specificitySc } from './specificity.js';
import { actionableScorer } from './actionable.js';

const GOOD_PLAN = `
## Overview
This system uses Redis for rate limiting with a sliding window algorithm.

## Architecture
We'll use a token bucket approach with Redis sorted sets. Each key maps to user:${new Date().toISOString()}.
Nginx sits in front, routes to the Node.js API layer.

## Implementation Steps
1. Set up Redis connection pool with 10 max connections
2. Implement \`checkRateLimit(userId, limit, windowMs)\` function
3. Use \`ZADD\` and \`ZREMRANGEBYSCORE\` for sliding window
4. Add middleware to Express: \`app.use(rateLimitMiddleware)\`
5. Configure per-IP fallback with Nginx \`limit_req_zone\`

\`\`\`typescript
const pipeline = redis.pipeline();
pipeline.zadd(key, now, requestId);
pipeline.zremrangebyscore(key, 0, now - windowMs);
pipeline.zcard(key);
pipeline.expire(key, Math.ceil(windowMs / 1000));
const [, , count] = await pipeline.exec();
\`\`\`

## Trade-offs & Risks
- Redis single point of failure — mitigate with Redis Sentinel
- Memory usage grows with request volume — set TTL strictly
- Distributed clock skew under 100ms acceptable for our use case
`;

const VAGUE_PLAN = `
## Overview
We might consider using some kind of caching solution. Perhaps Redis could work, or maybe another approach.

## Architecture
We could potentially look at various options. One might consider a distributed approach, though another solution might also be viable.

## Implementation
We should perhaps think about this more carefully. Maybe we'll figure it out.
`;

const evalCase = { task: 'Design a rate limiting system' };

describe('Coverage scorer', async () => {
  it('scores a good plan highly', async () => {
    const score = await coverageScorer.score(GOOD_PLAN, evalCase);
    assert.ok(score >= 0.6, `Expected score >= 0.6, got ${score}`);
  });

  it('scores a plan with missing sections lower', async () => {
    const partial = '## Overview\nSome content.\n## Architecture\nSome arch.';
    const score = await coverageScorer.score(partial, evalCase);
    assert.ok(score < 1.0, `Expected score < 1.0, got ${score}`);
  });

  it('returns 0-1 range', async () => {
    const score = await coverageScorer.score(GOOD_PLAN, evalCase);
    assert.ok(score >= 0 && score <= 1, `Score out of range: ${score}`);
  });
});

describe('Specificity scorer', async () => {
  it('scores concrete plan higher than vague plan', async () => {
    const goodScore = await specificitySc.score(GOOD_PLAN, evalCase);
    const vagueScore = await specificitySc.score(VAGUE_PLAN, evalCase);
    assert.ok(
      goodScore > vagueScore,
      `Expected good (${goodScore}) > vague (${vagueScore})`
    );
  });

  it('returns 0-1 range', async () => {
    const score = await specificitySc.score(GOOD_PLAN, evalCase);
    assert.ok(score >= 0 && score <= 1, `Score out of range: ${score}`);
  });
});

describe('Actionable scorer', async () => {
  it('scores plan with numbered steps and code blocks highly', async () => {
    const score = await actionableScorer.score(GOOD_PLAN, evalCase);
    assert.ok(score >= 0.5, `Expected score >= 0.5, got ${score}`);
  });

  it('scores plan without steps lower', async () => {
    const noSteps = '## Overview\nUse Redis.\n## Architecture\nPut Redis in front.';
    const score = await actionableScorer.score(noSteps, evalCase);
    assert.ok(score <= 0.5, `Expected score <= 0.5, got ${score}`);
  });

  it('returns 0-1 range', async () => {
    const score = await actionableScorer.score(GOOD_PLAN, evalCase);
    assert.ok(score >= 0 && score <= 1, `Score out of range: ${score}`);
  });
});
