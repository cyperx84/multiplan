package eval

import (
	"testing"
)

const goodPlan = `
## Overview
This system uses Redis for rate limiting with a sliding window algorithm.

## Architecture
We'll use a token bucket approach with Redis sorted sets. Each key maps to user:timestamp.
Nginx sits in front, routes to the Node.js API layer.

## Implementation Steps
1. Set up Redis connection pool with 10 max connections
2. Implement checkRateLimit(userId, limit, windowMs) function
3. Use ZADD and ZREMRANGEBYSCORE for sliding window
4. Add middleware to Express: app.use(rateLimitMiddleware)
5. Configure per-IP fallback with Nginx limit_req_zone

` + "```" + `typescript
const pipeline = redis.pipeline();
pipeline.zadd(key, now, requestId);
pipeline.zremrangebyscore(key, 0, now - windowMs);
pipeline.zcard(key);
pipeline.expire(key, Math.ceil(windowMs / 1000));
const [, , count] = await pipeline.exec();
` + "```" + `

## Trade-offs & Risks
- Redis single point of failure — mitigate with Redis Sentinel
- Memory usage grows with request volume — set TTL strictly
- Distributed clock skew under 100ms acceptable for our use case
`

const vaguePlan = `
## Overview
We might consider using some kind of caching solution. Perhaps Redis could work, or maybe another approach.

## Architecture
We could potentially look at various options. One might consider a distributed approach, though another solution might also be viable.

## Implementation
We should perhaps think about this more carefully. Maybe we'll figure it out.
`

func TestCoverageScorer(t *testing.T) {
	scorer := &CoverageScorer{}
	evalCase := &EvalCase{Task: "Design a rate limiting system"}

	t.Run("scores good plan highly", func(t *testing.T) {
		score, err := scorer.Score(goodPlan, evalCase)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if score < 0.6 {
			t.Errorf("expected score >= 0.6, got %.2f", score)
		}
	})

	t.Run("scores plan with missing sections lower", func(t *testing.T) {
		partial := "## Overview\nSome content.\n## Architecture\nSome arch."
		score, err := scorer.Score(partial, evalCase)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if score >= 1.0 {
			t.Errorf("expected score < 1.0, got %.2f", score)
		}
	})

	t.Run("returns 0-1 range", func(t *testing.T) {
		score, err := scorer.Score(goodPlan, evalCase)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if score < 0 || score > 1 {
			t.Errorf("score out of range: %.2f", score)
		}
	})
}

func TestSpecificityScorer(t *testing.T) {
	scorer := &SpecificityScorer{}
	evalCase := &EvalCase{Task: "Design a rate limiting system"}

	t.Run("scores concrete plan higher than vague plan", func(t *testing.T) {
		goodScore, err := scorer.Score(goodPlan, evalCase)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		vagueScore, err := scorer.Score(vaguePlan, evalCase)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if goodScore <= vagueScore {
			t.Errorf("expected good (%.2f) > vague (%.2f)", goodScore, vagueScore)
		}
	})

	t.Run("returns 0-1 range", func(t *testing.T) {
		score, err := scorer.Score(goodPlan, evalCase)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if score < 0 || score > 1 {
			t.Errorf("score out of range: %.2f", score)
		}
	})
}

func TestActionableScorer(t *testing.T) {
	scorer := &ActionableScorer{}
	evalCase := &EvalCase{Task: "Design a rate limiting system"}

	t.Run("scores plan with numbered steps and code blocks highly", func(t *testing.T) {
		score, err := scorer.Score(goodPlan, evalCase)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if score < 0.5 {
			t.Errorf("expected score >= 0.5, got %.2f", score)
		}
	})

	t.Run("scores plan without steps lower", func(t *testing.T) {
		noSteps := "## Overview\nUse Redis.\n## Architecture\nPut Redis in front."
		score, err := scorer.Score(noSteps, evalCase)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if score > 0.5 {
			t.Errorf("expected score <= 0.5, got %.2f", score)
		}
	})

	t.Run("returns 0-1 range", func(t *testing.T) {
		score, err := scorer.Score(goodPlan, evalCase)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if score < 0 || score > 1 {
			t.Errorf("score out of range: %.2f", score)
		}
	})
}
