---
title: The toolbox over a fake read API behind a real Connect server
status: active
sourceFile: internal/views/app/answer/toolbox_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'D-228 as amended (the Ask box and the MCP server offer a tool for exact decimal arithmetic): the design session''s brief, 2026-11-21'
  recorded_at: '2026-11-21T14:10:00.000Z'
id: TEST-views-answer-calc-toolbox
type: test
---
The toolbox over a fake read API behind a real Connect server: the calc tool lists with the Ask box's tools, a call makes only the settings read for the owner's number format, shows the expression with its exact result and the figures-given wording in the owner's decimal comma, refuses 1/0 with its reason as a card, and a token the read API does not accept is refused like any tool's.