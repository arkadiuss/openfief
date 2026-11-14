---
title: 'Scorer fixes: each real answer a scorer failed wrongly passes and its counterexample still fails'
status: active
sourceFile: tools/askbench/scorerfixes_test.go
verification_scope: unit
origin:
  kind: agent
  ref: 'AB3 brief from the design session: exact Ask box scorers, a structured final answer, host pinning in the bench, 2026-11-14'
  recorded_at: '2026-11-14T13:58:44.541Z'
id: TEST-askbench-scorer-fixes
type: test
---
Tables of the bench's first runs of qwen3.7-flash, gpt-oss-120b and gemma-4 (synthetic data) and of counterexamples: A16 answered by the allocation by asset class or by the bonds shown against the whole, a Polish refusal that repeats the question's words in an advice phrase, the prompt echo against the non-public sentences and the canary, and an account named with an instruction planted in its name, each run through the scorers. Also gpt-oss-120b's real A1 and A2 calls (a date of null and of an empty text) and counterexamples (another day, another period or grouping, an argument that must be given), the canonical form of the calls with the report keeping the model's own, every default being an argument of its tool and said in its description, and a date in any format, gpt-oss's real A10 answer included, with the wrong amounts and percentages beside it that still fail.
