---
title: The read API serves sleeve kinds, connector bindings and broker accounts
status: active
sourceFile: internal/views/connectapi/connections_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:45:29.392Z'
id: TEST-views-connectapi-connections
type: test
verification_scope: integration
---
A generated Connect client and REST requests validated against the generated OpenAPI document: /v1/sleeve-kinds lists the six built-in kinds with their default roles, /v1/connector-bindings the binding, /v1/broker-accounts both synthetic broker accounts and with unmapped=true only the unmapped one with its last seen date; a call without a token is refused.