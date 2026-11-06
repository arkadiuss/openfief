---
title: A browser cannot choose the trace
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T03:07:00.531Z'
id: SCEN-telemetry-trace-boundary
type: scenario
---
Given a request to the core's browser port with a forged traceparent, when the core serves it and proxies it to the view plugin, then its span starts a new trace linked to the forged one, and the plugin receives the core's trace context, never the browser's.