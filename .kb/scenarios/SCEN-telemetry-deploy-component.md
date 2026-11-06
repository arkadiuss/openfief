---
title: The tracing component turns export on for an installation
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T03:07:37.639Z'
id: SCEN-telemetry-deploy-component
type: scenario
---
Given an overlay of the bases with deploy/tracing and the owner's collector, when it is built, then the core and the operator read the tracing ConfigMap and the optional headers Secret and the core may reach the collector's OTLP ports; the bases alone set no tracing variable and reach no collector.