---
title: Table tests of the bench's arithmetic checks
status: active
sourceFile: tools/askbench/arithmetic_test.go
verification_scope: unit
origin:
  kind: agent
  ref: 'D-228 as amended (the Ask box and the MCP server offer a tool for exact decimal arithmetic): the design session''s brief, 2026-11-21'
  recorded_at: '2026-11-21T14:10:00.000Z'
id: TEST-askbench-arithmetic
type: test
---
Table tests of the bench's arithmetic checks: the amounts of the 30/70 question computed by code from the allocation under both readings of the total and refused when wrong, rounded or missing; the exactness of 0.3 against a float's 0.30000000000000004 and a long decimal in full or cut short; the only-numbers check letting no amount through; the before-taxes-and-fees wording; and a figure only a calc card holds counting as shown while the model's own sum does not.