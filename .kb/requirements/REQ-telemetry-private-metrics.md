---
title: 'Metrics leave a process only through an allow-list: with a metrics endpoint set (OTEL_EXPORTER_OTLP_METRICS_ENDPOINT, or the base endpoint with /v1/metrics, over OTLP/HTTP only, off with OTEL_METRICS_EXPORTER=none or OTEL_SDK_DISABLED=true), the core exports every minute only the five named instruments, each data point with only the attributes kind, browser, code, path and to whose string values are at most 64 characters, the scope with its name alone, the resource with its service name alone and no exemplar, and makes no call when nothing is left.'
status: open
origin:
  kind: agent
  ref: 'Demo operations brief from the design session: metrics reach Grafana through the core''s allow-list, plugins send them to the core and never to a collector (D-172, D-227, GR-12), 2026-11-19'
  recorded_at: '2026-11-19T17:14:04.064Z'
semantic_text: 'Metrics leave a process only through an allow-list: with a metrics endpoint set (OTEL_EXPORTER_OTLP_METRICS_ENDPOINT, or the base endpoint with /v1/metrics, over OTLP/HTTP only, off with OTEL_METRICS_EXPORTER=none or OTEL_SDK_DISABLED=true), the core exports every minute only the five named instruments, each data point with only the attributes kind, browser, code, path and to whose string values are at most 64 characters, the scope with its name alone, the resource with its service name alone and no exemplar, and makes no call when nothing is left.'
logic_claims:
  - CLAIM-B1804326A9C4A510
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: default
semantic_source_hash: 245e61d6605707669f50e1d7f65940de900e3fe61b1fd82dccc4abbc04c81ec9
semantic_inventory:
  - claim_key: CLAIM-B1804326A9C4A510
    claim_text: 'Metrics leave a process only through an allow-list: with a metrics endpoint set (OTEL_EXPORTER_OTLP_METRICS_ENDPOINT, or the base endpoint with /v1/metrics, over OTLP/HTTP only, off with OTEL_METRICS_EXPORTER=none or OTEL_SDK_DISABLED=true), the core exports every minute only the five named instruments, each data point with only the attributes kind, browser, code, path and to whose string values are at most 64 characters, the scope with its name alone, the resource with its service name alone and no exemplar, and makes no call when nothing is left'
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 553
id: REQ-telemetry-private-metrics
type: req
---
Metrics leave a process only through an allow-list: with a metrics endpoint set (OTEL_EXPORTER_OTLP_METRICS_ENDPOINT, or the base endpoint with /v1/metrics, over OTLP/HTTP only, off with OTEL_METRICS_EXPORTER=none or OTEL_SDK_DISABLED=true), the core exports every minute only the five named instruments, each data point with only the attributes kind, browser, code, path and to whose string values are at most 64 characters, the scope with its name alone, the resource with its service name alone and no exemplar, and makes no call when nothing is left.

## Context

D-172 says metrics go out over OTLP through the same allow-list as spans, and D-227 makes the UI's two client-event meters the first. A metric's attributes are words of a fixed vocabulary, never text a person typed, so a new metric or attribute needs a reviewed change to the allow-list and its test (GR-12).

## Source

> Route the UI's two D-227 meters, openfief.ui.client_events and openfief.ui.turnstile.token_seconds, that way, and export them through D-172's allow-list (the design session's demo operations brief, 2026-11-19).