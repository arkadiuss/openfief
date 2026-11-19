---
title: The telemetry collector's scrub is in every pipeline before the exporter and removes visitors' personal data from every kind of record
status: active
origin:
  kind: agent
  ref: 'Demo operations brief from the design session: ship the demo''s telemetry to Grafana Cloud with no visitor personal data (GR-12, D-172, D-203, D-227), 2026-11-19'
  recorded_at: '2026-11-19T16:20:41.988Z'
id: SCEN-demo-telemetry-scrub
type: scenario
---
Given the demo's telemetry layer, when the demo is validated, then every collector pipeline runs transform/scrub with nothing but the memory limiter before it and exports to otlphttp/grafana alone, the scrub fails closed and has a statement dropping keys and one rewriting values for each context of each signal, for the log body and for the span's name and status message, and a pipeline without the scrub, a second exporter, a scrub that fails open, a missing context, a key or a pattern left out each fail naming the problem; and when each statement is run on sample records, the keys that carry a client address, a session or token ID, a credential, a header or a question are dropped, the addresses, token prefixes, credentials and long random strings are rewritten, and times, versions, routes and UUIDs are left alone.