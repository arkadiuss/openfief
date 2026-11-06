---
title: Connector syncs, statement uploads, bond-lot entries, ledger imports, revaluations and operator reconciles must each run in a span named for the operation.
status: open
text_ref: docs/implementation-questions.md#tracing
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-implementation-questions-md-tracing
  - lane:strict
semantic_text: Connector syncs, statement uploads, bond-lot entries, ledger imports, revaluations and operator reconciles must each run in a span named for the operation.
logic_claims:
  - CLAIM-D54C8B01D161EA62
semantic_clauses:
  - Connector syncs, statement uploads, bond-lot entries, ledger imports, revaluations and operator reconciles must each run in a span named for the operation.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fec6ebb9d5a7216bbaf36a1224a8fb5056c4071811606aa09a0733528844a084
semantic_inventory:
  - claim_key: CLAIM-D54C8B01D161EA62
    claim_text: Connector syncs, statement uploads, bond-lot entries, ledger imports, revaluations and operator reconciles must each run in a span named for the operation
    role: normative
    status: modeled
    span:
      start: 0
      end: 154
origin:
  kind: agent
  recorded_at: '2026-11-02T03:07:24.227Z'
id: REQ-telemetry-domain-spans
type: req
---
Connector syncs, statement uploads, bond-lot entries, ledger imports, revaluations and operator reconciles must each run in a span named for the operation.

## Context

The repository owner asked for it directly ("add tracing and make pushing to otel configurable"). No decision in docs/DESIGN.md covers observability, so this is the narrowest sound version; docs/implementation-questions.md#tracing (items 71 to 74) parks the choices it leaves open for the author. Modeled by Claude Code; the author has not reviewed it yet. HTTP and query spans alone don't say which business operation was slow. The names are ingest.sync, ingest.import_file, ingest.import_bond_lots, ledger.import, valuation.revalue and operator.reconcile; a failure marks the span as an error without the error's message.

## Source

> add tracing and make pushing to otel configurable

The repository owner's request, 2026-11-02, relayed by the orchestrator; docs/implementation-questions.md#tracing
