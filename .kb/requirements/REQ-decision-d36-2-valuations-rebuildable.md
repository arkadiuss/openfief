---
title: Stored valuations must be rebuildable from the transaction log, prices, FX rates and index readings at any time.
status: open
text_ref: design-decisions:D-36#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-36-2
  - lane:strict
semantic_text: Stored valuations must be rebuildable from the transaction log, prices, FX rates and index readings at any time.
logic_claims:
  - CLAIM-24EFFE409F069E47
semantic_clauses:
  - Stored valuations must be rebuildable from the transaction log, prices, FX rates and index readings at any time.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8187a9956fb1724ef5c7011d52bb7bf2ba8cb9c25a3861145af3301d328edfc3
semantic_inventory:
  - claim_key: CLAIM-24EFFE409F069E47
    claim_text: Stored valuations must be rebuildable from the transaction log, prices, FX rates and index readings at any time
    role: normative
    status: modeled
    span:
      start: 0
      end: 111
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-36#2
  recorded_at: '2026-10-28T06:29:10.772Z'
id: REQ-decision-d36-2-valuations-rebuildable
type: req
---
Stored valuations must be rebuildable from the transaction log, prices, FX rates and index readings at any time.

## Context

The author decided D-36 in docs/DESIGN.md section 3.3; this requirement imports its claim D-36#2 as a firm rule, so changing it needs the author. Reason not stated in the decision. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Valuations are stored rows derived from the transaction log, prices, FX rates and index readings, and can be rebuilt from them at any time.

Source: OpenFief author decisions (docs/DESIGN.md section 3.3) - D-36#2
