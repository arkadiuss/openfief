---
title: The collector's scrub statements, run on sample records, drop and rewrite what they must and leave the rest, in every context, and every pipeline runs the scrub before its one exporter
status: active
origin:
  kind: agent
  ref: 'Demo operations brief from the design session: ship the demo''s telemetry to Grafana Cloud with no visitor personal data (GR-12, D-172, D-203, D-227), 2026-11-19'
  recorded_at: '2026-11-19T16:20:40.246Z'
sourceFile: internal/demo/scrub_test.go
verification_scope: unit
id: TEST-demo-telemetry-scrub
type: test
---
Reads the statements of transform/scrub out of deploy/demo/telemetry/collector.yaml, fails on a form of statement it does not understand, and applies delete_matching_keys, replace_all_patterns and replace_pattern with Go's regexp, the collector's own engine, to sample records in every context: the keys of client addresses, sessions, tokens, credentials, headers and questions are dropped and ordinary keys kept; IPv4 and IPv6 addresses in every form, the openfief token prefixes, Bearer and Basic credentials, JSON web tokens and long random strings are rewritten in attribute values, the log body and the span's name and status message, and times, versions, routes and UUIDs are not; every context carries the same patterns; every pipeline runs transform/scrub with only the memory limiter before it and exports to the one exporter, and the scrub fails closed. It does not run OTTL: the real collector was run on the same kinds of records when the layer was written.