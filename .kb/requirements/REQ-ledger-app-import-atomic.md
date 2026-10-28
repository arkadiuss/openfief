---
title: An import must append its transactions, store its assertion checks and run its import hooks in one database transaction that commits whole or not at all, and must report per account the earliest date whose current transactions it changed.
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: An import must append its transactions, store its assertion checks and run its import hooks in one database transaction that commits whole or not at all, and must report per account the earliest date whose current transactions it changed.
logic_claims:
  - CLAIM-ACF447E69493DF9D
semantic_clauses:
  - An import must append its transactions, store its assertion checks and run its import hooks in one database transaction that commits whole or not at all, and must report per account the earliest date whose current transactions it changed.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e00a968b5346775143e9fa0568c8ea788483ef13181fa50a5ccd22b022704b40
semantic_inventory:
  - claim_key: CLAIM-ACF447E69493DF9D
    claim_text: An import must append its transactions, store its assertion checks and run its import hooks in one database transaction that commits whole or not at all, and must report per account the earliest date whose current transactions it changed
    role: normative
    status: modeled
    span:
      start: 0
      end: 237
origin:
  kind: agent
  recorded_at: '2026-10-28T06:29:21.490Z'
id: REQ-ledger-app-import-atomic
type: req
---
An import must append its transactions, store its assertion checks and run its import hooks in one database transaction that commits whole or not at all, and must report per account the earliest date whose current transactions it changed.

## Context

D-36 makes valuations stored rows: every input change, a ledger import included, enqueues a revaluation in the same database transaction through River, and pending revaluations coalesce per account from the earliest affected date. The orchestrator asked for a transaction-scoped seam (Store.InTx, ImportHook) and for the summary to carry that date; River itself is a follow-up. A superseded version counts at its own account and date, so a correction moved to another account or date revalues both. Modeled by Claude Code for the M0 ledger-store task; the author has not reviewed it yet.

## Source

> Valuations are stored rows derived from the transaction log, prices, FX rates and index readings, and can be rebuilt from them at any time. Any change to those inputs enqueues a revaluation in the same database transaction, through a PostgreSQL-backed job queue (River). Pending revaluations coalesce per account from the earliest affected date.

docs/DESIGN.md D-36
