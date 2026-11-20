---
title: 'A view plugin''s metrics reach a collector only through the core: the plugin calls openfief.StartMetrics, which sends its metrics redacted every minute as OTLP/HTTP with its projected token to POST /otlp/v1/metrics on the core''s plugin port; the core takes the report only from a view plugin its TokenReview proves (401 without a token, 403 for any other caller, 415 for a body that is not protobuf, 413 over 256 KiB, 400 for no report), filters it again with the same allow-list, names its service after the plugin the token proved and never after the plugin''s own word, forwards what is left to the core''s own metrics endpoint or drops it when the core exports none, and answers 200 whether the collector took it or not.'
status: open
origin:
  kind: agent
  ref: 'Demo operations brief from the design session: metrics reach Grafana through the core''s allow-list, plugins send them to the core and never to a collector (D-172, D-227, GR-12), 2026-11-19'
  recorded_at: '2026-11-19T17:14:04.468Z'
semantic_text: 'A view plugin''s metrics reach a collector only through the core: the plugin calls openfief.StartMetrics, which sends its metrics redacted every minute as OTLP/HTTP with its projected token to POST /otlp/v1/metrics on the core''s plugin port; the core takes the report only from a view plugin its TokenReview proves (401 without a token, 403 for any other caller, 415 for a body that is not protobuf, 413 over 256 KiB, 400 for no report), filters it again with the same allow-list, names its service after the plugin the token proved and never after the plugin''s own word, forwards what is left to the core''s own metrics endpoint or drops it when the core exports none, and answers 200 whether the collector took it or not.'
logic_claims:
  - CLAIM-81C517E89F59CD19
  - CLAIM-8218BEC2015D6A2D
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: default
semantic_source_hash: cf0d074ca1c933c931a3a935d0a726b9adcc90b8cc55e090671ef47a2f673868
semantic_inventory:
  - claim_key: CLAIM-81C517E89F59CD19
    claim_text: 'A view plugin''s metrics reach a collector only through the core: the plugin calls openfief.StartMetrics, which sends its metrics redacted every minute as OTLP/HTTP with its projected token to POST /otlp/v1/metrics on the core''s plugin port'
    role: descriptive
    status: missing
    span:
      start: 0
      end: 239
  - claim_key: CLAIM-8218BEC2015D6A2D
    claim_text: the core takes the report only from a view plugin its TokenReview proves (401 without a token, 403 for any other caller, 415 for a body that is not protobuf, 413 over 256 KiB, 400 for no report), filters it again with the same allow-list, names its service after the plugin the token proved and never after the plugin's own word, forwards what is left to the core's own metrics endpoint or drops it when the core exports none, and answers 200 whether the collector took it or not
    role: normative
    status: ontology_gap
    span:
      start: 241
      end: 720
id: REQ-plugin-metrics-through-core
type: req
---
A view plugin's metrics reach a collector only through the core: the plugin calls openfief.StartMetrics, which sends its metrics redacted every minute as OTLP/HTTP with its projected token to POST /otlp/v1/metrics on the core's plugin port; the core takes the report only from a view plugin its TokenReview proves (401 without a token, 403 for any other caller, 415 for a body that is not protobuf, 413 over 256 KiB, 400 for no report), filters it again with the same allow-list, names its service after the plugin the token proved and never after the plugin's own word, forwards what is left to the core's own metrics endpoint or drops it when the core exports none, and answers 200 whether the collector took it or not.

## Context

D-172 has plugins send their telemetry to the core, which applies the allow-list and exports it, and no plugin reaches a collector. Spans are not built this way yet; metrics are, for the UI's client events. The plugin's NetworkPolicy already allows the core and nothing else, so the path needs no new rule.

## Source

> Extend the plugin to core path that spans use to metrics, so plugins send metrics to the core and the core exports them through D-172's allow-list; plugins must still never reach a collector themselves (the design session's demo operations brief, 2026-11-19).