---
title: Valuation must count each sale's gain for tax as its proceeds less its fees at the NBP rate of the day before the sale less its consumed lots' costs at the rate of the day before each was opened, and sum a tax year's gains through a day.
status: open
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

D-128 offsets the tax year's realized gains; D-20 converts tax amounts at the NBP rate of the day before. Buy fees stay out of the cost as D-34 has it. Modeled by Claude Code for the taxes work package; the author has not reviewed it yet.

## Source

> **[D-128]** A holding's net and liquidation values carry its own tax as if it alone were sold. Account and portfolio totals count the tax as PIT-38 does: the tax year's realized and unrealized gains and losses of every taxable account offset each other, and the base and the tax are rounded to whole złoty; a losses-offset line reconciles the rows with the total. Losses carried from earlier years are not counted. IKE and IKZE follow their own rules: an IKE's early return is taxed on what is withdrawn above what was paid in.

docs/DESIGN.md#D-128
