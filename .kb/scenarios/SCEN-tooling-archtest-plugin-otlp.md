---
title: A plugin, a view plugin's package or SDK code outside its telemetry package that imports an OpenTelemetry exporter or provider is reported, and the SDK's telemetry package, the core and a test's in-memory reader are not
status: active
origin:
  kind: agent
  ref: 'Demo operations brief from the design session: metrics reach Grafana through the core''s allow-list, plugins send them to the core and never to a collector (D-172, D-227, GR-12), 2026-11-19'
  recorded_at: '2026-11-19T17:14:03.926Z'
id: SCEN-tooling-archtest-plugin-otlp
type: scenario
---
The architecture test loads a repository where plugins/ui and a test of it import an exporter, internal/views/ui and internal/views/ghostfolio import a provider and an exporter, sdk/go/openfief imports an exporter, and sdk/go/telemetry, cmd/core, a plugin importing only the API and a plugin's test importing the SDK's in-memory reader do the same without being reported; it reports the five imports.