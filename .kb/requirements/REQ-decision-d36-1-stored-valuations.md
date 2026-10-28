---
title: Valuations must be stored rows derived from the transaction log, prices, FX rates and index readings.
status: open
text_ref: design-decisions:D-36#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-36-1
  - lane:strict
semantic_text: Valuations must be stored rows derived from the transaction log, prices, FX rates and index readings.
logic_claims:
  - CLAIM-C7F48192BFB45A12
semantic_clauses:
  - Valuations must be stored rows derived from the transaction log, prices, FX rates and index readings.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5615eb1c5630b4b49e15193a5c1a8afbb612857b17342efa114f22e309c3d0bf
semantic_inventory:
  - claim_key: CLAIM-C7F48192BFB45A12
    claim_text: Valuations must be stored rows derived from the transaction log, prices, FX rates and index readings
    role: normative
    status: modeled
    span:
      start: 0
      end: 100
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-36#1
  recorded_at: '2026-10-28T06:29:00.689Z'
id: REQ-decision-d36-1-stored-valuations
type: req
---
Valuations must be stored rows derived from the transaction log, prices, FX rates and index readings.

## Context

The author decided D-36 in docs/DESIGN.md section 3.3; this requirement imports its claim D-36#1 as a firm rule, so changing it needs the author. Reason not stated in the decision. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Valuations are stored rows derived from the transaction log, prices, FX rates and index readings, and can be rebuilt from them at any time.

Source: OpenFief author decisions (docs/DESIGN.md section 3.3) - D-36#1
