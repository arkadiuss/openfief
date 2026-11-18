---
title: The toolbox offers the Ask box's tools and runs each with the caller's own token
status: active
sourceFile: internal/views/ui/toolbox_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'D-213 with D-212, D-211, D-209 and D-203: the MCP server brief from the design session, 2026-11-18'
  recorded_at: '2026-11-18T09:00:00.000Z'
id: TEST-ui-toolbox
type: test
---
Serves the fake read API over an HTTP test server: the toolbox lists askTools' names, descriptions and schemas, reads only with the token it was given, writes the card as the model's tool result in the owner's number format, and ends unknown tools, refused arguments and a token the read API refuses as such, quoting no argument.