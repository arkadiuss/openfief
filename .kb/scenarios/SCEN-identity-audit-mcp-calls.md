---
title: An agent's call is one usage entry and shows nothing when it can't be entered
status: active
origin:
  kind: agent
  ref: 'D-213 with D-212, D-211, D-209 and D-203: the MCP server brief from the design session, 2026-11-18'
  recorded_at: '2026-11-18T09:00:00.000Z'
id: SCEN-identity-audit-mcp-calls
type: scenario
---
Given an agent calling tools, then each call is one audit entry of the action mcp_call whose actor is the script token by its ID, naming the tool, ok when the tool ran, refused when its arguments were not accepted and failed when its data could not be read, with no argument or result in it, and when the log can't write the entry the agent gets an error and none of the card.