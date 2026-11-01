---
title: Cross-check prices dropped because the connector's manifest doesn't declare them must raise an attention item.
status: open
text_ref: design-decisions:D-141#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-141-2
  - lane:strict
semantic_text: Cross-check prices dropped because the connector's manifest doesn't declare them must raise an attention item.
logic_claims:
  - CLAIM-19B274DA5F15DF0B
semantic_clauses:
  - Cross-check prices dropped because the connector's manifest doesn't declare them must raise an attention item.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bc99dc8034562dc83a09491f8fa7c9f0f4697e518a08924bc517311cc452ceeb
semantic_inventory:
  - claim_key: CLAIM-19B274DA5F15DF0B
    claim_text: Cross-check prices dropped because the connector's manifest doesn't declare them must raise an attention item
    role: normative
    status: modeled
    span:
      start: 0
      end: 109
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-141#2
  recorded_at: '2026-11-01T12:24:38.116Z'
id: REQ-ingest-undeclared-cross-check-prices-attention-item
type: req
---
Cross-check prices dropped because the connector's manifest doesn't declare them must raise an attention item.

## Context

The author recorded it on 2026-11-01, answering the implementation's question #25. The item goes to D-130's attention store (REQ-core-attention-item-store), so the owner learns that a connector sends prices it didn't declare. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Cross-check prices from a connector whose manifest doesn't declare them are dropped and raise an attention item (D-130); the rest of its delivery is imported.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-141#2
