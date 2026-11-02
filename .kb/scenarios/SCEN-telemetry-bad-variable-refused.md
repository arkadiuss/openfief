---
title: A bad tracing variable stops the start without quoting it
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T03:06:32.421Z'
id: SCEN-telemetry-bad-variable-refused
type: scenario
---
Given an endpoint without an http or https scheme, a protocol other than grpc or http/protobuf, headers that are not key=value pairs, an unknown sampler or a ratio outside 0 to 1, an exporter other than otlp or none, or a disabled flag that is not a boolean, when the configuration is parsed, then it fails with the error for that variable, whose text never contains the header or endpoint value.