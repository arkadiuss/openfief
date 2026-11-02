---
title: Valuation must count each sale's gain for tax as its proceeds less its fees at the NBP rate of the day before the sale less its consumed lots' costs and buy fees at the rate of the day before each was opened, count a held lot's buy fees in its tax cost the same way, count only sales as realized, and sum a tax year's gains through a day.
status: open
text_ref: docs/DESIGN.md#D-150
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-150
  - lane:strict
semantic_text: Valuation must count each sale's gain for tax as its proceeds less its fees at the NBP rate of the day before the sale less its consumed lots' costs and buy fees at the rate of the day before each was opened, count a held lot's buy fees in its tax cost the same way, count only sales as realized, and sum a tax year's gains through a day.
logic_claims:
  - CLAIM-8AD1ADB0E6BAA25B
semantic_clauses:
  - Valuation must count each sale's gain for tax as its proceeds less its fees at the NBP rate of the day before the sale less its consumed lots' costs and buy fees at the rate of the day before each was opened, count a held lot's buy fees in its tax cost the same way, count only sales as realized, and sum a tax year's gains through a day.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1c0bbb3003e625046244640b127b4a233cbf6b658b5b214a1a8d4d9b67c1d7b2
semantic_inventory:
  - claim_key: CLAIM-8AD1ADB0E6BAA25B
    claim_text: Valuation must count each sale's gain for tax as its proceeds less its fees at the NBP rate of the day before the sale less its consumed lots' costs and buy fees at the rate of the day before each was opened, count a held lot's buy fees in its tax cost the same way, count only sales as realized, and sum a tax year's gains through a day
    role: normative
    status: modeled
    span:
      start: 0
      end: 337
origin:
  kind: agent
  recorded_at: '2026-11-01T20:12:26.048Z'
id: REQ-valuation-realized-gains-buy-fees
type: req
---
Valuation must count each sale's gain for tax as its proceeds less its fees at the NBP rate of the day before the sale less its consumed lots' costs and buy fees at the rate of the day before each was opened, count a held lot's buy fees in its tax cost the same way, count only sales as realized, and sum a tax year's gains through a day.

## Context

D-150 puts a sale's buy fees in its tax cost, in each row's tax and in the totals alike, and keeps dividends out of the netting. Modeled by Claude Code for the taxes package D-144..D-150; the author has not reviewed it yet.

## Why this changed

It supersedes REQ-valuation-realized-gains: D-150 puts the buy fees in the tax cost of a sale, in each row's tax and in the totals alike, replacing the narrow reading that counted the lots' cost without them. The superseded requirement is closed.

## Source

> **[D-150]** The tax cost of a sale includes its buy fees, as PIT-38 deducts them, in each row's tax and in the totals alike. Dividends stay out of the netting, because Polish tax counts them as a separate source of income that capital losses don't offset.

docs/DESIGN.md#D-150
