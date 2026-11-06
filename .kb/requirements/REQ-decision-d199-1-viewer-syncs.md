---
title: Once viewers exist, a viewer must see only the syncs, sync attempts and import activity of the accounts they may view, and anything tied to a connector instance and no account must stay the owner's.
status: open
text_ref: design-decisions:D-199#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-199-1
  - lane:strict
semantic_text: Once viewers exist, a viewer must see only the syncs, sync attempts and import activity of the accounts they may view, and anything tied to a connector instance and no account must stay the owner's.
logic_claims:
  - CLAIM-1A9651199096EE46
semantic_clauses:
  - Once viewers exist, a viewer must see only the syncs, sync attempts and import activity of the accounts they may view, and anything tied to a connector instance and no account must stay the owner's.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b47ff40aca11b1055ed523b7076b7b771a78758b13e15096bc9f13dff31a7b3c
semantic_inventory:
  - claim_key: CLAIM-1A9651199096EE46
    claim_text: Once viewers exist, a viewer must see only the syncs, sync attempts and import activity of the accounts they may view, and anything tied to a connector instance and no account must stay the owner's
    role: normative
    status: modeled
    span:
      start: 0
      end: 197
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-199#1
  recorded_at: '2026-11-06T06:15:19.948Z'
id: REQ-decision-d199-1-viewer-syncs
type: req
---
Once viewers exist, a viewer must see only the syncs, sync attempts and import activity of the accounts they may view, and anything tied to a connector instance and no account must stay the owner's.

## Context

Author decision D-199 (UI v2 design session), extending D-165's rule for attention items to syncs and imports.

## Source

> **[D-199]** Once viewers exist (D-65), a viewer sees the syncs, sync attempts and import activity of the accounts they may view; anything tied to a connector instance and no account stays the owner's, as D-165 has for attention items.

docs/DESIGN.md#D-199
