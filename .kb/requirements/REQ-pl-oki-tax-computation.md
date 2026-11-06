---
title: The Polish pack must compute OKI's yearly tax from each OKI account's days held, averaging each class's end-of-day values plus its share of the payments counted up to each day's withdrawals over the days held, half up to grosze, exempting savings assets first then investment assets up to the year's amounts, and taxing the rest at the year's rate half up to grosze, with no rate or tax while the year's rate or exempt amounts are not known.
status: closed
text_ref: docs/DESIGN.md#D-127
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-127
  - lane:strict
semantic_text: The Polish pack must compute OKI's yearly tax from each OKI account's days held, averaging each class's end-of-day values plus its share of the payments counted up to each day's withdrawals over the days held, half up to grosze, exempting savings assets first then investment assets up to the year's amounts, and taxing the rest at the year's rate half up to grosze, with no rate or tax while the year's rate or exempt amounts are not known.
logic_claims:
  - CLAIM-4B7BFD6A08E3D1FC
semantic_clauses:
  - The Polish pack must compute OKI's yearly tax from each OKI account's days held, averaging each class's end-of-day values plus its share of the payments counted up to each day's withdrawals over the days held, half up to grosze, exempting savings assets first then investment assets up to the year's amounts, and taxing the rest at the year's rate half up to grosze, with no rate or tax while the year's rate or exempt amounts are not known.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 93a6e1a569a54bb74fe8016f9634529329b72bbefe916a4b126accee4e5392cf
semantic_inventory:
  - claim_key: CLAIM-4B7BFD6A08E3D1FC
    claim_text: The Polish pack must compute OKI's yearly tax from each OKI account's days held, averaging each class's end-of-day values plus its share of the payments counted up to each day's withdrawals over the days held, half up to grosze, exempting savings assets first then investment assets up to the year's amounts, and taxing the rest at the year's rate half up to grosze, with no rate or tax while the year's rate or exempt amounts are not known
    role: exception
    status: modeled
    span:
      start: 0
      end: 440
origin:
  kind: agent
  recorded_at: '2026-10-31T18:33:47.300Z'
id: REQ-pl-oki-tax-computation
type: req
---
The Polish pack must compute OKI's yearly tax from each OKI account's days held, averaging each class's end-of-day values plus its share of the payments counted up to each day's withdrawals over the days held, half up to grosze, exempting savings assets first then investment assets up to the year's amounts, and taxing the rest at the year's rate half up to grosze, with no rate or tax while the year's rate or exempt amounts are not known.

## Context

Closed: superseded by REQ-pl-oki-yearly-tax-exact. D-144 keeps each account's average exact and rounds the base and the tax to whole złoty instead of grosze, D-145 takes later rates and exempt amounts from the minister's announcements instead of computing them, and D-146 marks the current year an estimate. Modeled by Claude Code for the taxes package D-144..D-150; the author has not reviewed it yet.
