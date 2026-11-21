---
title: Askbench arithmetic graded by code
status: active
origin:
  kind: agent
  ref: 'D-228 as amended (the Ask box and the MCP server offer a tool for exact decimal arithmetic): the design session''s brief, 2026-11-21'
  recorded_at: '2026-11-21T14:10:00.000Z'
id: SCEN-askbench-arithmetic-graded-by-code
type: scenario
---
Given the bench's allocation, when an answer states the amounts to move to 30/70 stocks/bonds with or without the cash, before taxes and fees, then it passes, and an amount that differs, one amount only, or no amount fails; 0.3 passes and 0.30000000000000004 fails for 0.1 plus 0.2.