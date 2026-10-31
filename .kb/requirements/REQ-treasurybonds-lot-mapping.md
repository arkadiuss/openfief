---
title: The treasury bond connector must map each valid lot bought on or before the account's go-live to an opening lot on go-live and each later lot to a deposit and a buy on its purchase date, keyed by series and purchase date, map each deleted lot to deletions of those keys, and deliver each series' terms once.
status: open
text_ref: docs/DESIGN.md#D-98
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-98
  - lane:strict
semantic_text: The treasury bond connector must map each valid lot bought on or before the account's go-live to an opening lot on go-live and each later lot to a deposit and a buy on its purchase date, keyed by series and purchase date, map each deleted lot to deletions of those keys, and deliver each series' terms once.
logic_claims:
  - CLAIM-7A5D33813CE84588
semantic_clauses:
  - The treasury bond connector must map each valid lot bought on or before the account's go-live to an opening lot on go-live and each later lot to a deposit and a buy on its purchase date, keyed by series and purchase date, map each deleted lot to deletions of those keys, and deliver each series' terms once.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 42d55ec46c58c0a9243eee9e8c92db71f3e69e769d469659274c05ad30cd4227
semantic_inventory:
  - claim_key: CLAIM-7A5D33813CE84588
    claim_text: The treasury bond connector must map each valid lot bought on or before the account's go-live to an opening lot on go-live and each later lot to a deposit and a buy on its purchase date, keyed by series and purchase date, map each deleted lot to deletions of those keys, and deliver each series' terms once
    role: normative
    status: modeled
    span:
      start: 0
      end: 306
origin:
  kind: agent
  recorded_at: '2026-10-29T01:14:20.029Z'
id: REQ-treasurybonds-lot-mapping
type: req
---
The treasury bond connector must map each valid lot bought on or before the account's go-live to an opening lot on go-live and each later lot to a deposit and a buy on its purchase date, keyed by series and purchase date, map each deleted lot to deletions of those keys, and deliver each series' terms once.

## Context

D-98 replaces the earlier always deposit and buy mapping; D-100 adds deletion. Booking a lot bought on the go-live date itself as an opening lot is this package's narrow reading. Recorded by the bonds implementation agent (impl/bonds-2).

## Source

> A treasury bond lot bought before go-live is an opening lot at go-live, keeping its own purchase date and cost (D-23). A manually entered account may receive its opening lots one at a time, and each is idempotent per series and purchase date. A lot bought after go-live is a deposit of its cost and a buy on its purchase date.

docs/DESIGN.md#D-98
