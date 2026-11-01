---
title: Cross-check prices from a connector whose manifest doesn't declare them must be dropped.
status: open
text_ref: design-decisions:D-141#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-141-1
  - lane:strict
semantic_text: Cross-check prices from a connector whose manifest doesn't declare them must be dropped.
logic_claims:
  - CLAIM-88BAF9F926CBA7C6
semantic_clauses:
  - Cross-check prices from a connector whose manifest doesn't declare them must be dropped.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4841df208a9b384eccde9961d519068b6656060a150864ba6d639641599cb8db
semantic_inventory:
  - claim_key: CLAIM-88BAF9F926CBA7C6
    claim_text: Cross-check prices from a connector whose manifest doesn't declare them must be dropped
    role: normative
    status: modeled
    span:
      start: 0
      end: 87
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-141#1
  recorded_at: '2026-11-01T12:24:28.437Z'
id: REQ-ingest-undeclared-cross-check-prices-dropped
type: req
---
Cross-check prices from a connector whose manifest doesn't declare them must be dropped.

## Context

The author recorded it on 2026-11-01, answering the implementation's question #25. D-112 lets only a connector whose manifest declares it send cross-check prices (REQ-ingest-cross-check-prices-declared-only); this says what happens to the ones another connector sends. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Cross-check prices from a connector whose manifest doesn't declare them are dropped and raise an attention item (D-130); the rest of its delivery is imported.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-141#1
