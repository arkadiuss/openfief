---
title: Ui ask calc parser bounded
status: active
origin:
  kind: agent
  ref: 'D-228 as amended (the Ask box and the MCP server offer a tool for exact decimal arithmetic): the design session''s brief, 2026-11-21'
  recorded_at: '2026-11-21T14:10:00.000Z'
id: SCEN-ui-ask-calc-parser-bounded
type: scenario
---
Given an expression that is empty, 201 characters long, 25 parentheses deep, written with a function call, a letter, a decimal comma or scientific notation, or that divides by zero, when calc reads it, then it is refused with a reason that holds none of its text, and a code-injection string is only an unexpected character at a position.