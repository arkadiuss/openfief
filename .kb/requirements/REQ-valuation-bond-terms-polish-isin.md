---
title: Bond series terms must have a Polish (PL) ISIN.
status: open
text_ref: design-decisions:D-179#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-179-3
  - lane:strict
semantic_text: Bond series terms must have a Polish (PL) ISIN.
logic_claims:
  - CLAIM-380386277EC6C025
semantic_clauses:
  - Bond series terms must have a Polish (PL) ISIN.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 84a4790f7f6a0f7abbbb5ba27fd3f1f6beda2557e6a6c005c15dda1104ef8b0d
semantic_inventory:
  - claim_key: CLAIM-380386277EC6C025
    claim_text: Bond series terms must have a Polish (PL) ISIN
    role: normative
    status: modeled
    span:
      start: 0
      end: 46
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-179#3
  recorded_at: '2026-11-02T22:24:20.524Z'
id: REQ-valuation-bond-terms-polish-isin
type: req
---
Bond series terms must have a Polish (PL) ISIN.

## Context

The author decided it on 2026-11-02, after reviewing what a connector may deliver into the stores that value holdings (P0-5). Polish retail treasury bonds have an ISIN per series, and only they have series terms (D-9, D-70). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Only a connector whose signed manifest declares bond series terms, or market index levels, and whose grant approves that may deliver them; anything else it delivers of that kind is dropped with an attention item (D-130), as D-141 does for cross-check prices. Bond terms need a Polish (`PL`) ISIN, and every stored term and index level records the instance that delivered it.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-179#3
