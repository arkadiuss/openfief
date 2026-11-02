---
title: Shutdown sends the spans still buffered
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T03:06:41.184Z'
id: SCEN-telemetry-flush-on-shutdown
type: scenario
---
Given tracing pointed at a collector and spans ended but not yet due for a batch, when the process shuts its tracing down, then the collector receives those spans before shutdown returns, and nothing was sent before.