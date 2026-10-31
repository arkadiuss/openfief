---
title: The views list sleeve kinds, connector bindings and broker accounts
status: active
sourceFile: internal/views/app/connections_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:45:26.959Z'
id: TEST-views-app-connections
type: test
verification_scope: unit
---
Use-case tests of views app over fakes: SleeveKinds returns the stored kinds, ConnectorBindings the bindings, BrokerAccounts all or only the unmapped synthetic one; without the Connections port both lists are empty, and a failing store or ledger is passed on wrapped.