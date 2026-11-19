---
title: Metrics endpoint parsing, the metrics allow-list on exported and received metrics, and forwarding, in the SDK's telemetry package
status: active
origin:
  kind: agent
  ref: 'Demo operations brief from the design session: metrics reach Grafana through the core''s allow-list, plugins send them to the core and never to a collector (D-172, D-227, GR-12), 2026-11-19'
  recorded_at: '2026-11-19T17:13:58.103Z'
sourceFile: sdk/go/telemetry/metrics_test.go
verification_scope: integration
id: TEST-telemetry-metrics
type: test
---
ParseConfig takes the metrics endpoint from the standard variables and what turns it off; RedactMetrics keeps only allowed metrics, attributes and a bare resource and scope, drops exemplars, makes no call when nothing is left and returns the exporter's error; FilterOTLPMetrics does the same to a received report with the verified service name; the forwarder sends the filtered report with the configured headers to /v1/metrics against an httptest collector, names no endpoint in its error and drops on a nil forwarder.