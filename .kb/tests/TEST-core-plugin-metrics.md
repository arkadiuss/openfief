---
title: The core takes a view plugin's metrics report, filters and forwards it, and refuses what it must
status: active
origin:
  kind: agent
  ref: 'Demo operations brief from the design session: metrics reach Grafana through the core''s allow-list, plugins send them to the core and never to a collector (D-172, D-227, GR-12), 2026-11-19'
  recorded_at: '2026-11-19T17:13:58.140Z'
sourceFile: cmd/core/pluginmetrics_test.go
verification_scope: integration
id: TEST-core-plugin-metrics
type: test
---
Posts reports to the plugin port's handler in front of an httptest collector: 401 without a token, 403 for a caller that is no view plugin, 500 when the caller can't be checked, 415, 413 and 400 for bad bodies, 200 and a forward for good ones, the service name of the verified plugin, the allowed metric and attributes alone, nothing forwarded for a report of nothing allowed, a failing collector that doesn't fail the plugin and whose log line names no endpoint, a drop without an endpoint, and every other path left to the plugin port.