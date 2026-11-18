---
title: An agent reaches MCP only with the switch and toggle on and a token the owner allowed
status: active
origin:
  kind: agent
  ref: 'D-213 with D-212, D-211, D-209 and D-203: the MCP server brief from the design session, 2026-11-18'
  recorded_at: '2026-11-18T09:00:00.000Z'
id: SCEN-views-mcp-door
type: scenario
---
Given an MCP server whose switch and toggle the owner turns on and off, and script tokens of which one may use MCP, then with either off every request is 404 whatever its token, with no token or a token that is unknown, expired or revoked it is 401 with a Bearer challenge, with a valid token that may not use MCP it is 403, and with a token that may it speaks MCP over HTTP: initialize, tools/list and tools/call, with no tool list in any refusal, and the answers stop at the moment the toggle is turned off or the token revoked.