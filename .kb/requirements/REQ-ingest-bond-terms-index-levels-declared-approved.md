---
title: Only a connector whose signed manifest declares bond series terms, or market index levels, and whose grant approves that may deliver them.
status: open
text_ref: design-decisions:D-179#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-179-1
  - lane:strict
semantic_text: Only a connector whose signed manifest declares bond series terms, or market index levels, and whose grant approves that may deliver them.
logic_claims:
  - CLAIM-119E508DEB3C2E7B
semantic_clauses:
  - Only a connector whose signed manifest declares bond series terms, or market index levels, and whose grant approves that may deliver them.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c7d1591aba993530cc8329b364b4c6b227554032ebbc4f884c06461831b1516f
semantic_inventory:
  - claim_key: CLAIM-119E508DEB3C2E7B
    claim_text: Only a connector whose signed manifest declares bond series terms, or market index levels, and whose grant approves that may deliver them
    role: normative
    status: modeled
    span:
      start: 0
      end: 137
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-179#1
  recorded_at: '2026-11-02T22:24:04.946Z'
id: REQ-ingest-bond-terms-index-levels-declared-approved
type: req
---
Only a connector whose signed manifest declares bond series terms, or market index levels, and whose grant approves that may deliver them.

## Context

The author decided it on 2026-11-02, after reviewing what a connector may deliver into the stores that value holdings (P0-5). D-123 makes the market-data list a permission the grant approves, and D-112 lets only a connector whose manifest declares it send cross-check prices; D-179 applies the same to bond series terms, which value the owner's bonds, and to market index levels. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Only a connector whose signed manifest declares bond series terms, or market index levels, and whose grant approves that may deliver them; anything else it delivers of that kind is dropped with an attention item (D-130), as D-141 does for cross-check prices. Bond terms need a Polish (`PL`) ISIN, and every stored term and index level records the instance that delivered it.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-179#1
