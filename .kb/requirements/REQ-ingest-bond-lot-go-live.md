---
title: The core must send the account's go-live date with the owner's bond lots, the date it was seeded on or today when it is not seeded, and import the delivery as opening lots added one at a time.
status: open
text_ref: docs/DESIGN.md#D-98
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-98
  - lane:strict
semantic_text: The core must send the account's go-live date with the owner's bond lots, the date it was seeded on or today when it is not seeded, and import the delivery as opening lots added one at a time.
logic_claims:
  - CLAIM-3C1D129556E74C11
semantic_clauses:
  - The core must send the account's go-live date with the owner's bond lots, the date it was seeded on or today when it is not seeded, and import the delivery as opening lots added one at a time.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b343f516134753ff05d13c20cf9ebc4bc6065510052e37d08ac7d6cd510b0d35
semantic_inventory:
  - claim_key: CLAIM-3C1D129556E74C11
    claim_text: The core must send the account's go-live date with the owner's bond lots, the date it was seeded on or today when it is not seeded, and import the delivery as opening lots added one at a time
    role: normative
    status: modeled
    span:
      start: 0
      end: 191
origin:
  kind: agent
  recorded_at: '2026-10-30T06:33:36.320Z'
id: REQ-ingest-bond-lot-go-live
type: req
---
The core must send the account's go-live date with the owner's bond lots, the date it was seeded on or today when it is not seeded, and import the delivery as opening lots added one at a time.

## Context

D-98 decides the booking by go-live but not what go-live is for a manually entered account; taking the seeding date, or today for the first entry, is this package's narrow reading. Recorded by the bonds implementation agent (impl/bonds-2).

## Source

> A treasury bond lot bought before go-live is an opening lot at go-live, keeping its own purchase date and cost (D-23). A manually entered account may receive its opening lots one at a time, and each is idempotent per series and purchase date. A lot bought after go-live is a deposit of its cost and a buy on its purchase date.

docs/DESIGN.md#D-98
