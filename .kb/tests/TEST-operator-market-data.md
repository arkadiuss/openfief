---
title: The operator waits for the grant to approve market data, then records it
status: active
sourceFile: internal/plugins/operator/marketdata_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-01T10:05:24.518Z'
id: TEST-operator-market-data
type: test
---
TestReconcileMarketData runs the reconciler against envtest's API server: a Plugin whose manifest declares marketData with a PluginGrant that doesn't approve it has marketData pending approval and no status.effective.marketData; once the grant approves it the plugin runs and status.effective.marketData is true, which is what the core reads before telling it what the owner holds (D-123).