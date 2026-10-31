---
title: The views list installed plugins in name order and mark removed instances' transactions from that list
status: active
sourceFile: internal/views/app/plugins_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T03:29:22.273Z'
id: TEST-views-app-plugins
type: test
verification_scope: unit
---
Use-case tests over in-memory fakes: Views.Plugins sorts the port's records and maps their trust; without a Plugins port, or when it fails with ErrPluginsUnknown, Plugins fails with ErrPluginsUnknown and Transactions still answers, marking nothing removed; with the port, a transaction whose plugin:<instance> source names an instance that is no longer installed is marked removed, one still installed is not, and a non-connector source has no instance. A port failure other than unknown fails Plugins.