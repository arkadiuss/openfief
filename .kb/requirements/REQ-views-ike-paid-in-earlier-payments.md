---
title: What was paid into an IKE seeded from a starting balance must be the earlier payments the owner entered, dated on its go-live, plus its later deposits less its withdrawals, its seeding being no payment, and its liquidation total must be missing until the earlier payments are entered.
status: open
text_ref: docs/DESIGN.md#D-149
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-149
  - lane:strict
semantic_text: What was paid into an IKE seeded from a starting balance must be the earlier payments the owner entered, dated on its go-live, plus its later deposits less its withdrawals, its seeding being no payment, and its liquidation total must be missing until the earlier payments are entered.
logic_claims:
  - CLAIM-354B0EF0531AE968
semantic_clauses:
  - What was paid into an IKE seeded from a starting balance must be the earlier payments the owner entered, dated on its go-live, plus its later deposits less its withdrawals, its seeding being no payment, and its liquidation total must be missing until the earlier payments are entered.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c97a304ed17d60c00a206a3ec6b9b11d5bdef1c144b15b81ba60cdfc2e0ef30b
semantic_inventory:
  - claim_key: CLAIM-354B0EF0531AE968
    claim_text: What was paid into an IKE seeded from a starting balance must be the earlier payments the owner entered, dated on its go-live, plus its later deposits less its withdrawals, its seeding being no payment, and its liquidation total must be missing until the earlier payments are entered
    role: normative
    status: modeled
    span:
      start: 0
      end: 283
origin:
  kind: agent
  recorded_at: '2026-11-01T20:13:15.102Z'
id: REQ-views-ike-paid-in-earlier-payments
type: req
---
What was paid into an IKE seeded from a starting balance must be the earlier payments the owner entered, dated on its go-live, plus its later deposits less its withdrawals, its seeding being no payment, and its liquidation total must be missing until the earlier payments are entered.

## Context

D-149 taxes an IKE whose earlier history is unknown above the total payments in the owner entered. Showing the liquidation total as missing until then, as D-78 does for IKZE without the marginal rate, is a narrow reading. Modeled by Claude Code for the taxes package D-144..D-150; the author has not reviewed it yet.

## Source

> **[D-149]** For an IKE whose earlier history is unknown (D-23), the owner enters its total payments in, which its early return is taxed above.

docs/DESIGN.md#D-149
