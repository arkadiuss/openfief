---
title: Ask Box Calc Parser
status: active
fact_kind: subject
subject_key: ask_box.calc_parser
origin:
  kind: agent
  ref: 'D-228 as amended (the Ask box and the MCP server offer a tool for exact decimal arithmetic): the design session''s brief, 2026-11-21'
  recorded_at: '2026-11-21T14:10:00.000Z'
id: FACT-SUBJECT-6C74DF76ABC98C54
type: fact
---
The calc tool's parser: the fixed grammar it reads, evaluating nothing else, and the bounds on length, nesting, digits and division by zero it refuses with a reason that names no figure. It does not cover the arithmetic of a valid expression.