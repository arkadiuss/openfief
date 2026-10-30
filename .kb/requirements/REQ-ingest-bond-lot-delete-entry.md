---
title: The core must delete an owner's bond lot named by its series and purchase date through the connector bound to its account, which delivers a deletion of each transaction the lot was booked as, and refuse a lot never entered.
status: open
text_ref: docs/DESIGN.md#D-100
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-100
  - lane:strict
semantic_text: The core must delete an owner's bond lot named by its series and purchase date through the connector bound to its account, which delivers a deletion of each transaction the lot was booked as, and refuse a lot never entered.
logic_claims:
  - CLAIM-1472FEC4759F10D5
semantic_clauses:
  - The core must delete an owner's bond lot named by its series and purchase date through the connector bound to its account, which delivers a deletion of each transaction the lot was booked as, and refuse a lot never entered.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: af0098a007b3b45f96f155bcb60c1e3f49d820157d04cb17dabdf47e7810e268
semantic_inventory:
  - claim_key: CLAIM-1472FEC4759F10D5
    claim_text: The core must delete an owner's bond lot named by its series and purchase date through the connector bound to its account, which delivers a deletion of each transaction the lot was booked as, and refuse a lot never entered
    role: normative
    status: modeled
    span:
      start: 0
      end: 222
origin:
  kind: agent
  recorded_at: '2026-10-30T06:33:43.002Z'
id: REQ-ingest-bond-lot-delete-entry
type: req
---
The core must delete an owner's bond lot named by its series and purchase date through the connector bound to its account, which delivers a deletion of each transaction the lot was booked as, and refuse a lot never entered.

## Context

D-100 lets the owner delete a lot as a correction that supersedes it and erases nothing; D-49 sends entries through the core to the connector. Recorded by the bonds implementation agent (impl/bonds-2).

## Source

> Manual bond lots are cash purchases at 100 PLN or exchanges of maturing bonds at 99.90 PLN, whose 0.10 PLN discount is taxed with the interest on redemption. The owner can delete a lot as a correction, which supersedes it (D-18) and erases nothing.

docs/DESIGN.md#D-100
