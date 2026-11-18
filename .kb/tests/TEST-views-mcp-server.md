---
title: The MCP endpoint refuses before any tool, speaks MCP, runs tools as the caller's token and audits each call
status: active
sourceFile: internal/views/mcpserver/server_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'D-213 with D-212, D-211, D-209 and D-203: the MCP server brief from the design session, 2026-11-18'
  recorded_at: '2026-11-18T09:00:00.000Z'
id: TEST-views-mcp-server
type: test
---
Serves the endpoint over an HTTP test server with fake ports and speaks JSON-RPC to it, and through the official SDK's client: the status of each refusal with no tool list, the switch read at every request, initialize, tools/list and tools/call, each call's own token, the audit entry of each outcome, an unknown tool with no entry, a log that is down, and no stream or session.