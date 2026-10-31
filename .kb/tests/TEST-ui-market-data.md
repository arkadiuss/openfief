---
title: The settings page ranks market-data sources through the core
status: active
sourceFile: internal/views/ui/marketdata_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T18:00:48.027Z'
id: TEST-ui-market-data
type: test
verification_scope: integration
---
Handler tests of the built-in UI against a fake MarketDataService: the owner's settings page lists each kind's sources with their institution and default place and Up/Down buttons, an empty kind says so; a post moves a source up or down, keeps the top one, resets to the defaults with an empty order, and an unknown kind, no button and the core's refusal answer 400 with the reason.