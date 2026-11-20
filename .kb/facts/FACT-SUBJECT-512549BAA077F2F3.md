---
title: Tooling Archtest Plugin OTLP Imports
status: active
fact_kind: subject
subject_key: tooling.archtest_plugin_otlp_imports
origin:
  kind: agent
  ref: Author's decision to upgrade to kibi 2.45 and review the subjects, 2026-11-20
  recorded_at: '2026-11-20T12:51:08.000Z'
id: FACT-SUBJECT-512549BAA077F2F3
type: fact
---
The architecture tests that fail when plugin packages, a view plugin's own package or SDK code outside sdk/go/telemetry import an OpenTelemetry exporter.