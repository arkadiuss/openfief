---
title: Undeclared cross-check prices must raise one attention item per connector instance, with a detail that names no count or batch, and the instance's next delivery without undeclared ones must clear it.
status: open
text_ref: docs/DESIGN.md#D-141
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-141
  - lane:strict
semantic_text: Undeclared cross-check prices must raise one attention item per connector instance, with a detail that names no count or batch, and the instance's next delivery without undeclared ones must clear it.
logic_claims:
  - CLAIM-852CA6F2C6F66554
semantic_clauses:
  - Undeclared cross-check prices must raise one attention item per connector instance, with a detail that names no count or batch, and the instance's next delivery without undeclared ones must clear it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0a098972659d869e28eeea73cb5f5dad08633f7435d29b5f345baf1e95ab309d
semantic_inventory:
  - claim_key: CLAIM-852CA6F2C6F66554
    claim_text: Undeclared cross-check prices must raise one attention item per connector instance, with a detail that names no count or batch, and the instance's next delivery without undeclared ones must clear it
    role: normative
    status: modeled
    span:
      start: 0
      end: 198
origin:
  kind: agent
  recorded_at: '2026-11-02T13:17:49.340Z'
id: REQ-ingest-undeclared-cross-check-prices-item-per-instance
type: req
---
Undeclared cross-check prices must raise one attention item per connector instance, with a detail that names no count or batch, and the instance's next delivery without undeclared ones must clear it.

## Context

D-141 has undeclared cross-check prices dropped with an attention item (D-130) and the rest of the delivery imported. Which subject the item has and when it resolves are the implementation's reading: one item per instance, so repeated syncs don't pile up items or reopen one the owner acknowledged, resolved once the instance stops sending them or a version that declares them runs. The skipped ingest problem stays in the batch history too (REQ-ingest-cross-check-prices).

## Source

> **[D-141]** Cross-check prices from a connector whose manifest doesn't declare them are dropped and raise an attention item (D-130); the rest of its delivery is imported.

docs/DESIGN.md#D-141
