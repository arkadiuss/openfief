---
title: A token not allowed to use MCP is refused and entered
status: active
origin:
  kind: agent
  ref: the design session's answers on D-213, D-209, D-165 and D-199, relayed by the coordinator, 2026-11-18
  recorded_at: '2026-11-18T17:00:00.000Z'
id: SCEN-identity-audit-mcp-not-allowed
type: scenario
---
Given a valid token without the MCP flag, then it is answered 403 and one audit entry by that token, refused, names no tool; a request with no or a bad token is 401, writes no entry and is counted.