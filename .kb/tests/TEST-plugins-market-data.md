---
title: Only a declared and approved connector is a market-data connector
status: active
sourceFile: internal/plugins/marketdata_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T10:05:21.509Z'
id: TEST-plugins-market-data
type: test
---
TestMarketData runs a table of plugin and grant pairs through Manifest.Validate, Plugin.Requested, NeedsApproval and Effective: only a connector whose manifest declares marketData and whose grant approves it ends up with EffectivePermissions.MarketData, a declared but unapproved one waits for approval, and a view may not declare it. TestMarketDataLabelAndDiff reads the declaration from a manifest label, refuses it on a view, tells manifests apart by it and shows it in an upgrade's permissions diff (D-123).