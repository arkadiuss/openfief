---
title: Each sync request to a connector instance bound to accounts must carry, for each bound account once in ID order, its go-live date, tax wrapper, non-zero holdings and open lots today, and the stored series terms of its treasury bonds; an instance bound to no account gets none.
status: open
text_ref: docs/DESIGN.md#D-117
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-117
  - lane:strict
semantic_text: Each sync request to a connector instance bound to accounts must carry, for each bound account once in ID order, its go-live date, tax wrapper, non-zero holdings and open lots today, and the stored series terms of its treasury bonds; an instance bound to no account gets none.
logic_claims:
  - CLAIM-4512B4382550D09D
semantic_clauses:
  - Each sync request to a connector instance bound to accounts must carry, for each bound account once in ID order, its go-live date, tax wrapper, non-zero holdings and open lots today, and the stored series terms of its treasury bonds; an instance bound to no account gets none.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9308a7a5e2b1ebd28ba472461222ecf1a9185564f13066471b20d7b93133326d
semantic_inventory:
  - claim_key: CLAIM-4512B4382550D09D
    claim_text: Each sync request to a connector instance bound to accounts must carry, for each bound account once in ID order, its go-live date, tax wrapper, non-zero holdings and open lots today, and the stored series terms of its treasury bonds; an instance bound to no account gets none
    role: normative
    status: modeled
    span:
      start: 0
      end: 275
origin:
  kind: agent
  recorded_at: '2026-10-31T13:25:45.351Z'
id: REQ-ingest-sync-carries-holdings
type: req
---
Each sync request to a connector instance bound to accounts must carry, for each bound account once in ID order, its go-live date, tax wrapper, non-zero holdings and open lots today, and the stored series terms of its treasury bonds; an instance bound to no account gets none.

## Context

D-117 gives connectors the holdings and open lots of their bound accounts, as D-93 gives market-data connectors the instruments to price. The go-live date, tax wrapper and series terms are this implementation's narrow reading: a connector stores nothing, and computing a coupon and its tax needs them (implementation questions). Recorded by the implementation agent of impl/decisions-117; default firmness, not reviewed by the author yet.

## Source

> Each sync request to a connector carries the holdings and open lots of the accounts bound to that instance, as D-93's request carries the instruments to price. The treasury-bonds connector uses them to deliver each coupon that fell due since its last sync as interest (D-101), computed with the Polish pack's bond rules, so the core holds no bond logic.

docs/DESIGN.md#D-117
