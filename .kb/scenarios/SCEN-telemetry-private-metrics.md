---
title: The core exports only the allowed metrics with only the vocabulary's attributes, and nothing when nothing is allowed
status: active
origin:
  kind: agent
  ref: 'Demo operations brief from the design session: metrics reach Grafana through the core''s allow-list, plugins send them to the core and never to a collector (D-172, D-227, GR-12), 2026-11-19'
  recorded_at: '2026-11-19T17:14:03.539Z'
id: SCEN-telemetry-private-metrics
type: scenario
---
Given a metrics endpoint, when a counter carries a client address, a session and a question beside the words of the vocabulary and another counter is not on the list, then the collector receives the allowed counter with its kind and browser alone, a resource of the service name alone and no exemplar, and not the other counter; and a process with nothing allowed makes no call; the metrics endpoint comes from the standard variables, and the none exporter and OTEL_SDK_DISABLED turn it off.