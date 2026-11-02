---
title: A buy's fees must go to the lot it opens when it acquires one holding, and a lot consumed in part must split each fee with its units as it splits its cost, rounding the consumed share half-even to two places and keeping the exact rest.
status: open
text_ref: docs/DESIGN.md#D-150
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-150
  - lane:strict
semantic_text: A buy's fees must go to the lot it opens when it acquires one holding, and a lot consumed in part must split each fee with its units as it splits its cost, rounding the consumed share half-even to two places and keeping the exact rest.
logic_claims:
  - CLAIM-A5E57AC2F5F22E73
semantic_clauses:
  - A buy's fees must go to the lot it opens when it acquires one holding, and a lot consumed in part must split each fee with its units as it splits its cost, rounding the consumed share half-even to two places and keeping the exact rest.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 09131675427a4c49616169f4fefd36431768b54ec11bd658d60a6affe76f4555
semantic_inventory:
  - claim_key: CLAIM-A5E57AC2F5F22E73
    claim_text: A buy's fees must go to the lot it opens when it acquires one holding, and a lot consumed in part must split each fee with its units as it splits its cost, rounding the consumed share half-even to two places and keeping the exact rest
    role: normative
    status: modeled
    span:
      start: 0
      end: 234
origin:
  kind: agent
  recorded_at: '2026-11-01T20:12:38.127Z'
id: REQ-ledger-lot-buy-fees
type: req
---
A buy's fees must go to the lot it opens when it acquires one holding, and a lot consumed in part must split each fee with its units as it splits its cost, rounding the consumed share half-even to two places and keeping the exact rest.

## Context

D-150 needs each sale's buy fees for its tax cost while D-34 keeps fees out of lot cost, so a lot carries them apart. Giving a buy's fees to its lot only when it acquires one holding is a narrow reading. Modeled by Claude Code for the taxes package D-144..D-150; the author has not reviewed it yet.

## Source

> **[D-150]** The tax cost of a sale includes its buy fees, as PIT-38 deducts them, in each row's tax and in the totals alike. Dividends stay out of the netting, because Polish tax counts them as a separate source of income that capital losses don't offset.

docs/DESIGN.md#D-150
