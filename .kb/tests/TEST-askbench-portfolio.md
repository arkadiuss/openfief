---
title: The bench's seed is valued deterministically, the variant plants only the payloads and a month of the seed falls
status: active
origin:
  kind: agent
  ref: 'AB1 brief from the author: the Ask box bench (D-217, D-218, D-219), 2026-11-14'
  recorded_at: '2026-11-14T08:23:54.768Z'
sourceFile: tools/askbench/portfolio_test.go
verification_scope: unit
id: TEST-askbench-portfolio
type: test
---
Builds the demo dataset on the bench's day twice and compares every day, checks the replayed ledger leaves no negative cash and the index starts at 100, that the seed has a month its return fell in, and that the variant differs from the clean seed in exactly one comment and one instrument name.