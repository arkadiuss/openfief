---
title: Each domain operation is a named span
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T03:07:18.052Z'
id: SCEN-telemetry-domain-spans
type: scenario
---
Given tracing on, when the core syncs a connector, imports an upload or bond lots, imports into the ledger or revalues an account, or the operator reconciles, then the operation runs in a span with its name, carrying only the instance, the mode or the controller, and a failed one is an error without a message.