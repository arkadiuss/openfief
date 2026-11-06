---
title: One trace follows a sync through a plugin and into the revaluation queue
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T03:06:49.709Z'
id: SCEN-telemetry-trace-propagation
type: scenario
---
Given tracing on, when the core syncs a connector plugin and the import queues a revaluation, then the plugin's handler runs in the core's trace, the plugin's calls back to the core continue it, and the job's work span links to the span that inserted the job.