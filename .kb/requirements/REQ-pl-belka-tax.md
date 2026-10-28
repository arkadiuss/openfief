---
title: The net-if-redeemed value of a taxable bond lot must deduct Belka tax of 19% of accrued interest minus the early-redemption fee, rounded half up to whole grosze, and no tax when that is not positive.
status: open
text_ref: docs/DESIGN.md#D-25
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-25
  - lane:strict
semantic_text: The net-if-redeemed value of a taxable bond lot must deduct Belka tax of 19% of accrued interest minus the early-redemption fee, rounded half up to whole grosze, and no tax when that is not positive.
logic_claims:
  - CLAIM-A48205296C783414
semantic_clauses:
  - The net-if-redeemed value of a taxable bond lot must deduct Belka tax of 19% of accrued interest minus the early-redemption fee, rounded half up to whole grosze, and no tax when that is not positive.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 13ed8df0d9e0b4aea0b5908033783fc0258b8f63653d890d856d3eb47fae48ae
semantic_inventory:
  - claim_key: CLAIM-A48205296C783414
    claim_text: The net-if-redeemed value of a taxable bond lot must deduct Belka tax of 19% of accrued interest minus the early-redemption fee, rounded half up to whole grosze, and no tax when that is not positive
    role: normative
    status: modeled
    span:
      start: 0
      end: 198
origin:
  kind: agent
  recorded_at: '2026-10-28T05:22:08.401Z'
id: REQ-pl-belka-tax
type: req
---
The net-if-redeemed value of a taxable bond lot must deduct Belka tax of 19% of accrued interest minus the early-redemption fee, rounded half up to whole grosze, and no tax when that is not positive.

## Context

D-25 makes the default bond value net of Belka tax. The 19% rate is art. 30a ust. 1 pkt 2 of the PIT act; withheld tax on interest is rounded to whole grosze (art. 63 par. 1a of the tax ordinance); the fee lowers the taxable interest. The rounding direction (half up) is the agent's reading and is unverified. Refines REQ-AUTO-E10BE8CE13E67657. Modeled by Claude Code for the M0 valuation task; the author has not reviewed it yet. obligacjeskarbowe.pl and gov.pl were not reachable from the agent's sandbox, so the rule was checked only against search-result excerpts of issue letters and secondary calculators.

## Source

> the default is net-if-redeemed-today (after the early-redemption fee and Belka tax, no tax inside IKE)

docs/DESIGN.md D-25
