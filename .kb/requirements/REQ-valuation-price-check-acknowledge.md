---
title: The owner must be able to acknowledge a flagged price check, and each acknowledgement must be recorded append-only.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-112-2
  - lane:strict
text_ref: design-decisions:D-112#2
origin:
  kind: import
  ref: design-decisions:D-112#2
  recorded_at: '2026-10-29T10:55:41.089Z'
semantic_text: The owner must be able to acknowledge a flagged price check, and each acknowledgement must be recorded append-only.
logic_claims:
  - CLAIM-E9417C7D5B92F34F
semantic_clauses:
  - The owner must be able to acknowledge a flagged price check, and each acknowledgement must be recorded append-only.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 126e161605aad09ae04e5cf67aa923f76750ec7ae9f436878f941bd2bd871365
semantic_inventory:
  - claim_key: CLAIM-E9417C7D5B92F34F
    claim_text: The owner must be able to acknowledge a flagged price check, and each acknowledgement must be recorded append-only
    role: normative
    status: modeled
    span:
      start: 0
      end: 114
id: REQ-valuation-price-check-acknowledge
type: req
---
The owner must be able to acknowledge a flagged price check, and each acknowledgement must be recorded append-only.

## Context

D-110 flags a cross-check price that differs from the stored price by more than its threshold (REQ-valuation-price-check-one-percent). D-112 lets the owner acknowledge such a flag and keeps every acknowledgement, append-only. The author recorded D-109 to D-112 after the design session that answered the implementation's questions.

## Source

> Only a connector whose manifest declares it may send cross-check prices. The owner can acknowledge a flagged check, which is recorded append-only, and the UI shows the latest check per instrument with its history.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-112#2
