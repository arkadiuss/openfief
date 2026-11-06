---
title: The ingest store must keep with every completed import batch the accounts it booked into, and list them with the completed batches.
status: open
text_ref: docs/DESIGN.md#D-140
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-140
  - lane:strict
semantic_text: The ingest store must keep with every completed import batch the accounts it booked into, and list them with the completed batches.
logic_claims:
  - CLAIM-F66D3ABF93A55C66
semantic_clauses:
  - The ingest store must keep with every completed import batch the accounts it booked into, and list them with the completed batches.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8cf4347e982cf4afc33f750bc626e7b43ffe8de784517d25ea5a78d0fa30f15f
semantic_inventory:
  - claim_key: CLAIM-F66D3ABF93A55C66
    claim_text: The ingest store must keep with every completed import batch the accounts it booked into, and list them with the completed batches
    role: normative
    status: modeled
    span:
      start: 0
      end: 130
origin:
  kind: agent
  recorded_at: '2026-11-06T03:43:12.509Z'
id: REQ-ingest-batch-accounts
type: req
---
The ingest store must keep with every completed import batch the accounts it booked into, and list them with the completed batches.

## Context

The recent activity names the accounts an upload booked into (D-140, D-193), so the batch history keeps their internal IDs, never a broker account number (GR-12). Modeled by Claude Code for the UI v2 Inbox; the author has not reviewed the wording.

## Source

> **[D-140]** An upload of a statement covering several accounts runs in the mode chosen for the account it was uploaded for, which the owner can override: an opening upload seeds only the accounts not seeded yet, and a transactions upload books every mapped account.

docs/DESIGN.md#D-140
