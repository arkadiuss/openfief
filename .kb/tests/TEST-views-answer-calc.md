---
title: Table tests of answer.Calc
status: active
sourceFile: internal/views/app/answer/calc_test.go
verification_scope: unit
origin:
  kind: agent
  ref: 'D-228 as amended (the Ask box and the MCP server offer a tool for exact decimal arithmetic): the design session''s brief, 2026-11-21'
  recorded_at: '2026-11-21T14:10:00.000Z'
id: TEST-views-answer-calc
type: test
---
Table tests of answer.Calc: 0.1 + 0.2 is 0.3, a long decimal and a product of two 20-digit integers exact, precedence, parentheses, signs, percent signs, the multiplication and division signs, a quotient rounded half-even once at 20 places with ties to the even digit and flagged as rounded; and every refusal (empty, over 200 characters, nested over 20 levels, a number or a result of too many digits, division by zero, a decimal comma, a letter, a function call, scientific notation, a missing operand, an unclosed parenthesis, a code injection attempt) with a reason that quotes nothing; then the card. A false pass would need the wants computed by the test, but each is written out.