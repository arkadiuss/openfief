---
title: StartMetrics sends a plugin's redacted metrics to the core's address with the plugin's token and is off when asked
status: active
origin:
  kind: agent
  ref: 'Demo operations brief from the design session: metrics reach Grafana through the core''s allow-list, plugins send them to the core and never to a collector (D-172, D-227, GR-12), 2026-11-19'
  recorded_at: '2026-11-19T17:13:58.172Z'
sourceFile: sdk/go/openfief/metrics_test.go
verification_scope: integration
id: TEST-sdk-start-metrics
type: test
---
Starts the plugin's metrics against a fake core on an httptest server: a counted event reaches POST /otlp/v1/metrics with the bearer token, the instance as the service name and the vocabulary's attribute alone; OTEL_METRICS_EXPORTER=none sends nothing; a missing core address is an error.