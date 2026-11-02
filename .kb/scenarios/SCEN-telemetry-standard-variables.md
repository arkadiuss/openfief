---
title: The standard variables point export at the owner's collector
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T03:06:23.673Z'
id: SCEN-telemetry-standard-variables
type: scenario
---
Given OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_EXPORTER_OTLP_PROTOCOL, OTEL_EXPORTER_OTLP_HEADERS, OTEL_TRACES_SAMPLER and its argument, OTEL_SERVICE_NAME and OTEL_RESOURCE_ATTRIBUTES, when a process starts, then it exports over that protocol to that endpoint with those headers, samples as named and names its service and resource as given; the tracing component sets them for the core and the operator from a ConfigMap and a Secret.