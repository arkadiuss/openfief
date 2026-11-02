---
title: Without an endpoint nothing is recorded or sent
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T03:06:03.870Z'
id: SCEN-telemetry-off-by-default
type: scenario
---
Given the core, the operator or a plugin started without OTEL_EXPORTER_OTLP_ENDPOINT, or with OTEL_SDK_DISABLED=true, when it serves requests and runs jobs, then its spans do not record, no exporter is created and no connection to a collector is opened, while the deploy bases set no tracing variable.