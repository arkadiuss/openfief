---
title: The missing-settings check must raise an item for each account seeded from a starting balance in a wrapper that taxes what is withdrawn above what was paid in whose earlier payments the owner hasn't entered, its subject earlier-payments and the account.
status: open
text_ref: docs/DESIGN.md#D-149
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-149
  - lane:strict
semantic_text: The missing-settings check must raise an item for each account seeded from a starting balance in a wrapper that taxes what is withdrawn above what was paid in whose earlier payments the owner hasn't entered, its subject earlier-payments and the account.
logic_claims:
  - CLAIM-EFABD2551E20C907
semantic_clauses:
  - The missing-settings check must raise an item for each account seeded from a starting balance in a wrapper that taxes what is withdrawn above what was paid in whose earlier payments the owner hasn't entered, its subject earlier-payments and the account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3017fe227c65446be2d0537da0ee5c2cbf0af19f4e328c8a625c1db61bffa0d0
semantic_inventory:
  - claim_key: CLAIM-EFABD2551E20C907
    claim_text: The missing-settings check must raise an item for each account seeded from a starting balance in a wrapper that taxes what is withdrawn above what was paid in whose earlier payments the owner hasn't entered, its subject earlier-payments and the account
    role: normative
    status: modeled
    span:
      start: 0
      end: 252
origin:
  kind: agent
  recorded_at: '2026-11-01T20:13:27.539Z'
id: REQ-attention-earlier-payments-missing
type: req
---
The missing-settings check must raise an item for each account seeded from a starting balance in a wrapper that taxes what is withdrawn above what was paid in whose earlier payments the owner hasn't entered, its subject earlier-payments and the account.

## Context

D-130 has every check that needs the owner raise an attention item, including missing settings; D-149 needs the earlier payments of an IKE whose history is unknown. Modeled by Claude Code for the taxes package D-144..D-150; the author has not reviewed it yet.

## Source

> **[D-149]** For an IKE whose earlier history is unknown (D-23), the owner enters its total payments in, which its early return is taxed above.

docs/DESIGN.md#D-149
