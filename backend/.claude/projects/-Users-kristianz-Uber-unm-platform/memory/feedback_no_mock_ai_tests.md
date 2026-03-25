---
name: no_mock_ai_tests
description: AI tests must use real OpenAI — no mocking allowed
type: feedback
---

All AI tests must use the real OpenAI API. Do NOT mock OpenAI responses in tests.

**Why:** User explicitly stated this — "we cant mock AI we need to test the real thing". The `UNM_OPENAI_API_KEY` env var is set in `ai.env`.

**How to apply:**
- Tests that call OpenAI should use `NewOpenAIClient()` (reads real key from env)
- Tests should skip gracefully with `t.Skip` if `UNM_OPENAI_API_KEY` is not set
- Never use `httptest.NewServer` or mock HTTP servers to intercept OpenAI calls in production test paths
- The 6.10 AI Test Suite (115 questions) must run against real OpenAI
- 6.7 AI Advisor REST API tests should call real OpenAI — use integration test pattern with skip-if-no-key
