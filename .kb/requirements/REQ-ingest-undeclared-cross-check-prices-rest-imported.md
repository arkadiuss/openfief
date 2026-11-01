---
title: When a connector's undeclared cross-check prices are dropped, the rest of its delivery must be imported.
status: open
text_ref: design-decisions:D-141#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-141-3
  - lane:strict
semantic_text: When a connector's undeclared cross-check prices are dropped, the rest of its delivery must be imported.
logic_claims:
  - CLAIM-0F9442E0C2303DAB
semantic_clauses:
  - When a connector's undeclared cross-check prices are dropped, the rest of its delivery must be imported.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e81e1d06d77ec1f9fcd09e862551a9092eee95608ca6f450f59aad7f011769d5
semantic_inventory:
  - claim_key: CLAIM-0F9442E0C2303DAB
    claim_text: When a connector's undeclared cross-check prices are dropped, the rest of its delivery must be imported
    role: condition
    status: modeled
    span:
      start: 0
      end: 103
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-141#3
  recorded_at: '2026-11-01T12:24:46.841Z'
id: REQ-ingest-undeclared-cross-check-prices-rest-imported
type: req
---
When a connector's undeclared cross-check prices are dropped, the rest of its delivery must be imported.

## Context

The author recorded it on 2026-11-01, answering the implementation's question #25. Reason not stated. The rest of a delivery is its transactions, balance assertions and the other entries a connector sends (D-49). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Cross-check prices from a connector whose manifest doesn't declare them are dropped and raise an attention item (D-130); the rest of its delivery is imported.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-141#3
