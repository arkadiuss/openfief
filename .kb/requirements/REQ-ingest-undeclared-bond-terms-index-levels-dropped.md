---
title: Bond series terms or market index levels from a connector that may not deliver them must be dropped with an attention item, as undeclared cross-check prices are.
status: open
text_ref: design-decisions:D-179#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-179-2
  - lane:strict
semantic_text: Bond series terms or market index levels from a connector that may not deliver them must be dropped with an attention item, as undeclared cross-check prices are.
logic_claims:
  - CLAIM-FF2D757560C99A28
semantic_clauses:
  - Bond series terms or market index levels from a connector that may not deliver them must be dropped with an attention item, as undeclared cross-check prices are.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 32c61b8c8bb260b8bc9e114d716eba55ce589676abf796b164a1d7c714931b17
semantic_inventory:
  - claim_key: CLAIM-FF2D757560C99A28
    claim_text: Bond series terms or market index levels from a connector that may not deliver them must be dropped with an attention item, as undeclared cross-check prices are
    role: normative
    status: modeled
    span:
      start: 0
      end: 160
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-179#2
  recorded_at: '2026-11-02T22:24:11.632Z'
id: REQ-ingest-undeclared-bond-terms-index-levels-dropped
type: req
---
Bond series terms or market index levels from a connector that may not deliver them must be dropped with an attention item, as undeclared cross-check prices are.

## Context

The author decided it on 2026-11-02, after reviewing what a connector may deliver into the stores that value holdings (P0-5). D-141 drops cross-check prices a connector's manifest doesn't declare and raises an attention item (D-130), importing the rest of the delivery. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Only a connector whose signed manifest declares bond series terms, or market index levels, and whose grant approves that may deliver them; anything else it delivers of that kind is dropped with an attention item (D-130), as D-141 does for cross-check prices. Bond terms need a Polish (`PL`) ISIN, and every stored term and index level records the instance that delivered it.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-179#2
