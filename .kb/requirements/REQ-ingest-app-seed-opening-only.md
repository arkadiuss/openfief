---
title: An opening import batch must seed through the ledger's seeding use case, which refuses an already seeded account, and a transactions batch that holds opening lots or opening cash must be refused before anything is imported.
status: open
text_ref: docs/DESIGN.md#D-23
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-23
  - lane:strict
semantic_text: An opening import batch must seed through the ledger's seeding use case, which refuses an already seeded account, and a transactions batch that holds opening lots or opening cash must be refused before anything is imported.
logic_claims:
  - CLAIM-49C4C404000888F9
semantic_clauses:
  - An opening import batch must seed through the ledger's seeding use case, which refuses an already seeded account, and a transactions batch that holds opening lots or opening cash must be refused before anything is imported.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0fb5ef13513a939406a908d81ab1f855bba3cf93f13cb63dac462fd8eb00cc8d
semantic_inventory:
  - claim_key: CLAIM-49C4C404000888F9
    claim_text: An opening import batch must seed through the ledger's seeding use case, which refuses an already seeded account, and a transactions batch that holds opening lots or opening cash must be refused before anything is imported
    role: normative
    status: modeled
    span:
      start: 0
      end: 222
origin:
  kind: agent
  recorded_at: '2026-10-28T08:39:38.208Z'
id: REQ-ingest-app-seed-opening-only
type: req
---
An opening import batch must seed through the ledger's seeding use case, which refuses an already seeded account, and a transactions batch that holds opening lots or opening cash must be refused before anything is imported.

## Context

D-23 splits ingest into an opening mode that seeds at go-live and a transactions mode from go-live on. Routing the opening batch to Ledger.Seed is what enforces the author's refuse-re-seed decision at the ingest path. The author told the orchestrator, in answer to this task (not yet written into DESIGN.md): opening cash stays a deposit or withdrawal against the equity ledger account, and seeding an account that is already seeded is refused with a clear error. Refines REQ-AUTO-3C2A493C661181C7. Modeled by Claude Code for the M0 seed-guard task; the author has not reviewed it yet.

## Source

> Accounts are seeded from each broker's current lot-level positions (opening lots), with transactions from go-live.

docs/DESIGN.md D-23
