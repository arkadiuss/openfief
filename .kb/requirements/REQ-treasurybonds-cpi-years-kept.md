---
title: The treasury-bonds connector must keep each year of CPI that GUS served in full across syncs and ask GUS for at most four years it lacks per sync, so every coupon and maturity since each lot's purchase is booked within a few syncs however many years its lots span.
status: open
text_ref: docs/DESIGN.md#D-143
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-143
  - lane:strict
semantic_text: The treasury-bonds connector must keep each year of CPI that GUS served in full across syncs and ask GUS for at most four years it lacks per sync, so every coupon and maturity since each lot's purchase is booked within a few syncs however many years its lots span.
logic_claims:
  - CLAIM-AB9F564322F20A87
semantic_clauses:
  - The treasury-bonds connector must keep each year of CPI that GUS served in full across syncs and ask GUS for at most four years it lacks per sync, so every coupon and maturity since each lot's purchase is booked within a few syncs however many years its lots span.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 480f3450aac2582fdca78f7a601573a9529fd85305a23d0be623a19a13abca1f
semantic_inventory:
  - claim_key: CLAIM-AB9F564322F20A87
    claim_text: The treasury-bonds connector must keep each year of CPI that GUS served in full across syncs and ask GUS for at most four years it lacks per sync, so every coupon and maturity since each lot's purchase is booked within a few syncs however many years its lots span
    role: normative
    status: modeled
    span:
      start: 0
      end: 263
origin:
  kind: agent
  recorded_at: '2026-11-02T09:18:37.906Z'
id: REQ-treasurybonds-cpi-years-kept
type: req
---
The treasury-bonds connector must keep each year of CPI that GUS served in full across syncs and ask GUS for at most four years it lacks per sync, so every coupon and maturity since each lot's purchase is booked within a few syncs however many years its lots span.

## Context

D-143 books every coupon since each lot's purchase on every sync. A sync asks GUS for at most four years of CPI to stay within its anonymous limit of 5 requests a second and 100 in 15 minutes; without keeping the years already served, every sync would ask for the same earliest four and the coupons and maturities of later years would never be booked. Keeping them in the plugin's memory, not in storage, is this implementation's reading: the connector stores nothing (D-117).

## Source

> **[D-143]** Besides the holdings and open lots, a sync request carries each bound account's tax wrapper and the stored terms of the bond series among its lots. The treasury-bonds connector books every coupon due since each lot's purchase on every sync, keyed per lot and coupon date, so a sync stays idempotent and a lot entered late gets its earlier coupons.

docs/DESIGN.md#D-143
