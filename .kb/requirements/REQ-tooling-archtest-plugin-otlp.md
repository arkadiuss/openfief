---
title: The architecture tests fail when a package under plugins/, a view plugin's own package (internal/views/ui, internal/views/ghostfolio) or SDK code outside sdk/go/telemetry imports an OpenTelemetry exporter (tests included) or, in code that is no test, an OpenTelemetry SDK provider, so no plugin configures where telemetry goes and a plugin's metrics can reach the core alone.
status: open
origin:
  kind: agent
  ref: 'Demo operations brief from the design session: metrics reach Grafana through the core''s allow-list, plugins send them to the core and never to a collector (D-172, D-227, GR-12), 2026-11-19'
  recorded_at: '2026-11-19T17:14:04.913Z'
semantic_text: The architecture tests fail when a package under plugins/, a view plugin's own package (internal/views/ui, internal/views/ghostfolio) or SDK code outside sdk/go/telemetry imports an OpenTelemetry exporter (tests included) or, in code that is no test, an OpenTelemetry SDK provider, so no plugin configures where telemetry goes and a plugin's metrics can reach the core alone.
logic_claims:
  - CLAIM-870B00ACE83E3085
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: default
semantic_source_hash: 0e6909288b58abd054a391d1dba86a5f42ec83e06306a950f1a161066e7b5cc7
semantic_inventory:
  - claim_key: CLAIM-870B00ACE83E3085
    claim_text: The architecture tests fail when a package under plugins/, a view plugin's own package (internal/views/ui, internal/views/ghostfolio) or SDK code outside sdk/go/telemetry imports an OpenTelemetry exporter (tests included) or, in code that is no test, an OpenTelemetry SDK provider, so no plugin configures where telemetry goes and a plugin's metrics can reach the core alone
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 374
    subjects:
      - tooling.archtest_plugin_otlp_imports
id: REQ-tooling-archtest-plugin-otlp
type: req
---
The architecture tests fail when a package under plugins/, a view plugin's own package (internal/views/ui, internal/views/ghostfolio) or SDK code outside sdk/go/telemetry imports an OpenTelemetry exporter (tests included) or, in code that is no test, an OpenTelemetry SDK provider, so no plugin configures where telemetry goes and a plugin's metrics can reach the core alone.

## Context

The design session asked for a test that no plugin configures an OTLP exporter directly, as the other half of D-172's rule that plugins never reach a collector: the operator's pod spec gives plugins no OTEL_EXPORTER variable, and this keeps plugin code from adding an exporter of its own.

## Source

> Add an architecture or unit test that no plugin configures an OTLP exporter directly; the design session asked for this (the design session's demo operations brief, 2026-11-19).