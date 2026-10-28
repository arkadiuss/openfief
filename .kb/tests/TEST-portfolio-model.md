---
title: Model portfolio targets have a kind, a name, a weight in 0.01% steps, and weights summing to 100%
status: active
sourceFile: internal/portfolio/model_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T22:11:17.024Z'
id: TEST-portfolio-model
type: test
---
Table tests of Model.Validate over synthetic targets (XX test ISINs, the made-up index FAKE20): off and empty is the valid default; on needs targets; 60/40 and a 33.33/33.33/33.34 mix with an index pass; sums of 99.99 and 100.01, a third decimal place, zero and negative weights, the same target twice, an index without a currency, an instrument with one, no kind, no name, a name over 64 characters and 51 targets fail with ErrInvalidModel; exactly 50 targets pass. TestTargetKind round-trips the kind names. A false pass would need the sum check to accept a near miss, which the 99.99 and 100.01 cases rule out.