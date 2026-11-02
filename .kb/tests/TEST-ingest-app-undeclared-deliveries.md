---
title: Bond terms and index levels an instance may not deliver are dropped with one attention item
status: active
verification_scope: unit
sourceFile: internal/ingest/app/undeclared_test.go
origin:
  kind: agent
  recorded_at: '2026-11-02T22:24:35.064Z'
id: TEST-ingest-app-undeclared-deliveries
type: test
---
Syncs a fake connector through Connectors with in-memory stores, declarations and attention items. For an instance whose declarations allow neither, only bond terms or both, it checks which terms and levels are stored, each with the instance as its source, the skipped problems per kind added to the batch, the transactions still imported, and that the instance's one attention item is raised, keeps the same detail on a repeated drop and is cleared by a clean delivery. Storing a dropped value or a spoofed source fails it.