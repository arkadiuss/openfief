---
title: Ui ask calc exact
status: active
origin:
  kind: agent
  ref: 'D-228 as amended (the Ask box and the MCP server offer a tool for exact decimal arithmetic): the design session''s brief, 2026-11-21'
  recorded_at: '2026-11-21T14:10:00.000Z'
id: SCEN-ui-ask-calc-exact
type: scenario
---
Given the expressions 0.1+0.2, 0.1234567890123456789*3 and 200*5%, when calc computes them, then the results are exactly 0.3, 0.3703703670370370367 and 10; and 1/3 is 0.33333333333333333333 with a note that the division was rounded, while 1/8 is exactly 0.125 with none.