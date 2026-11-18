---
title: An agent's tools are the Ask box's read-only tools, read as the agent's own token
status: active
origin:
  kind: agent
  ref: 'D-213 with D-212, D-211, D-209 and D-203: the MCP server brief from the design session, 2026-11-18'
  recorded_at: '2026-11-18T09:00:00.000Z'
id: SCEN-views-mcp-tools-are-ask-box-tools
type: scenario
---
Given an agent with an allowed script token, then tools/list names exactly the Ask box's tools, each marked read-only, and each tools/call runs that tool's answerer against the read API with the agent's own token as bearer and nothing else, returns the card as JSON, and refuses an unknown tool or arguments the schema doesn't list without reading anything.