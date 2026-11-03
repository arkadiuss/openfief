---
title: Bond series terms or market index levels an instance may not deliver must raise one attention item per connector instance, with a detail that names only the kinds dropped, and the instance's next delivery that drops none must clear it.
status: open
text_ref: docs/DESIGN.md#D-179
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-179
  - lane:strict
semantic_text: Bond series terms or market index levels an instance may not deliver must raise one attention item per connector instance, with a detail that names only the kinds dropped, and the instance's next delivery that drops none must clear it.
logic_claims:
  - CLAIM-34239A35151E37EB
semantic_clauses:
  - Bond series terms or market index levels an instance may not deliver must raise one attention item per connector instance, with a detail that names only the kinds dropped, and the instance's next delivery that drops none must clear it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d41e4355f6f8595da993af3128ed7700566767545ee7bb5bb85cbe4b3a20852c
semantic_inventory:
  - claim_key: CLAIM-34239A35151E37EB
    claim_text: Bond series terms or market index levels an instance may not deliver must raise one attention item per connector instance, with a detail that names only the kinds dropped, and the instance's next delivery that drops none must clear it
    role: normative
    status: modeled
    span:
      start: 0
      end: 234
origin:
  kind: agent
  recorded_at: '2026-11-02T22:25:05.421Z'
id: REQ-ingest-undeclared-deliveries-item-per-instance
type: req
---
Bond series terms or market index levels an instance may not deliver must raise one attention item per connector instance, with a detail that names only the kinds dropped, and the instance's next delivery that drops none must clear it.

## Context

D-179 drops them with an attention item (D-130), as D-141 does for cross-check prices. Which subject the item has and when it resolves are the implementation's reading, the same as D-141's: one item per instance, so repeated syncs don't pile up items or reopen one the owner acknowledged, resolved once the instance stops delivering them or a version that declares them runs with its grant approving them. A skipped ingest problem per kind stays in the batch history too.

## Source

> Only a connector whose signed manifest declares bond series terms, or market index levels, and whose grant approves that may deliver them; anything else it delivers of that kind is dropped with an attention item (D-130), as D-141 does for cross-check prices. Bond terms need a Polish (`PL`) ISIN, and every stored term and index level records the instance that delivered it.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-179#2

