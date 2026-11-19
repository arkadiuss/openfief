---
title: A view plugin's metrics reach the collector through the core, filtered and named by the verified plugin, and no plugin reaches a collector
status: active
origin:
  kind: agent
  ref: 'Demo operations brief from the design session: metrics reach Grafana through the core''s allow-list, plugins send them to the core and never to a collector (D-172, D-227, GR-12), 2026-11-19'
  recorded_at: '2026-11-19T17:14:03.756Z'
id: SCEN-plugin-metrics-through-core
type: scenario
---
Given a view plugin that calls StartMetrics and a core that exports metrics, when the plugin counts an event with an address beside the vocabulary's words, then the core is sent a protobuf report with the plugin's token, the collector receives the allowed metric with the vocabulary's attributes alone under the service name of the plugin the token proved, and not the metric nobody allowed; a report without a token, from a caller that is no view plugin, that is not protobuf, over the cap or no report is refused; a collector that is down does not fail the plugin; without a metrics endpoint the report is taken and dropped.