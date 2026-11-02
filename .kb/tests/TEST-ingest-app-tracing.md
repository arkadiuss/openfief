---
title: Syncs, uploads and bond-lot entries run in named spans that name only the instance and mode
status: active
sourceFile: internal/ingest/app/tracing_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-02T03:05:49.739Z'
id: TEST-ingest-app-tracing
type: test
---
With a span recorder as the global tracer provider, a sync, a statement upload and a failing bond-lot entry through app.Connectors over the package's fakes produce ingest.sync, ingest.import_file and ingest.import_bond_lots spans carrying exactly the instance and, for the first two, the mode; the failed one has an error status without a description. Spans of other tests are told apart by the instance.