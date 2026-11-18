---
title: MCP is not served where grants are ignored
status: active
origin:
  kind: agent
  ref: the design session's answers on D-213, D-209, D-165 and D-199, relayed by the coordinator, 2026-11-18
  recorded_at: '2026-11-18T17:00:00.000Z'
id: SCEN-views-mcp-needs-grant-enforcement
type: scenario
---
Given the switch and the toggle on and an authorizer that does not enforce a token's grants, then every request is 503, and with either off every request is 404 whatever the authorizer.