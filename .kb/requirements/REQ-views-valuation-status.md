---
title: The read API must report, for each stored account, the last date its valuations are complete for, absent until it has been valued, and whether a revaluation is pending and whether one is running.
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: The read API must report, for each stored account, the last date its valuations are complete for, absent until it has been valued, and whether a revaluation is pending and whether one is running.
logic_claims:
  - CLAIM-6CE2C360D5EA773E
semantic_clauses:
  - The read API must report, for each stored account, the last date its valuations are complete for, absent until it has been valued, and whether a revaluation is pending and whether one is running.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: dd4eeb041d24286ae4806efe82206cf4a46e7e8c5a6f9c306ebec6184a95d4b2
semantic_inventory:
  - claim_key: CLAIM-6CE2C360D5EA773E
    claim_text: The read API must report, for each stored account, the last date its valuations are complete for, absent until it has been valued, and whether a revaluation is pending and whether one is running
    role: normative
    status: modeled
    span:
      start: 0
      end: 194
origin:
  kind: agent
  recorded_at: '2026-10-28T07:26:04.134Z'
id: REQ-views-valuation-status
type: req
---
The read API must report, for each stored account, the last date its valuations are complete for, absent until it has been valued, and whether a revaluation is pending and whether one is running.

## Context

Refines the firm requirement REQ-decision-d36-6-api-valuation-status (D-36 #6) for the views context: it reads the status through a consumer-side ValuationStatuses port that the valuation context's job queue implements. The account comes from the ledger, so an unknown account is refused and every stored account appears in the list. Modeled by Claude Code for the M0 read API task; the author has not reviewed it yet.

## Source

> Valuations are stored rows derived from the transaction log, prices, FX rates and index readings, and can be rebuilt from them at any time. [...] The read API reports, per account, the date it is valued through and whether a revaluation is pending or running, and the UI shows it.

docs/DESIGN.md D-36
