---
title: Valuation must count each sale's gain for tax as its proceeds less its fees at the NBP rate of the day before the sale less its consumed lots' costs at the rate of the day before each was opened, and sum a tax year's gains through a day.
status: closed
text_ref: docs/DESIGN.md#D-128
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-128
  - lane:strict
semantic_text: Valuation must count each sale's gain for tax as its proceeds less its fees at the NBP rate of the day before the sale less its consumed lots' costs at the rate of the day before each was opened, and sum a tax year's gains through a day.
logic_claims:
  - CLAIM-01DAAC961BD42C0E
semantic_clauses:
  - Valuation must count each sale's gain for tax as its proceeds less its fees at the NBP rate of the day before the sale less its consumed lots' costs at the rate of the day before each was opened, and sum a tax year's gains through a day.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a3b118b01741ed3cdf98ee82173bc9ed5ae328ff6bcc6ab019befa7e72bf2f64
semantic_inventory:
  - claim_key: CLAIM-01DAAC961BD42C0E
    claim_text: Valuation must count each sale's gain for tax as its proceeds less its fees at the NBP rate of the day before the sale less its consumed lots' costs at the rate of the day before each was opened, and sum a tax year's gains through a day
    role: normative
    status: modeled
    span:
      start: 0
      end: 236
origin:
  kind: agent
  recorded_at: '2026-10-31T18:34:17.210Z'
id: REQ-valuation-realized-gains
type: req
---
Valuation must count each sale's gain for tax as its proceeds less its fees at the NBP rate of the day before the sale less its consumed lots' costs at the rate of the day before each was opened, and sum a tax year's gains through a day.

## Context

Closed: superseded by REQ-valuation-realized-gains-buy-fees. D-150 puts the buy fees in the tax cost of a sale, in each row's tax and in the totals alike, replacing the narrow reading that counted the lots' cost without them. Modeled by Claude Code for the taxes package D-144..D-150; the author has not reviewed it yet.
