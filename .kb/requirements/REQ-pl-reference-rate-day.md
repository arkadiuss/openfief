---
title: The NBP reference rate that sets a ROR or DOR interest period must be the rate in force on the tenth business day before the first day of the month the period starts in, where business days exclude Saturdays and the statutory days off work of that year.
status: open
text_ref: docs/DESIGN.md#D-9
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-9
  - lane:strict
semantic_text: The NBP reference rate that sets a ROR or DOR interest period must be the rate in force on the tenth business day before the first day of the month the period starts in, where business days exclude Saturdays and the statutory days off work of that year.
logic_claims:
  - CLAIM-CF10D3E4086EE986
semantic_clauses:
  - The NBP reference rate that sets a ROR or DOR interest period must be the rate in force on the tenth business day before the first day of the month the period starts in, where business days exclude Saturdays and the statutory days off work of that year.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 67a5c0f71d8dc42eddc2a69a2f8ae9efae66000c7d7e7a7f3d0b07fee90d0a67
semantic_inventory:
  - claim_key: CLAIM-CF10D3E4086EE986
    claim_text: The NBP reference rate that sets a ROR or DOR interest period must be the rate in force on the tenth business day before the first day of the month the period starts in, where business days exclude Saturdays and the statutory days off work of that year
    role: normative
    status: modeled
    span:
      start: 0
      end: 252
origin:
  kind: agent
  recorded_at: '2026-10-29T03:45:57.373Z'
id: REQ-pl-reference-rate-day
type: req
---
The NBP reference rate that sets a ROR or DOR interest period must be the rate in force on the tenth business day before the first day of the month the period starts in, where business days exclude Saturdays and the statutory days off work of that year.

## Context

D-9 decides that bond valuation is computed; the ROR and DOR issue letters (ust. 16, 29 and 33, Załącznik nr 1) fix which reference rate applies and what a business day is. The holiday list follows art. 1 of the act on days off work, year by year from 1991; the epidemic day off of art. 1a is not modeled, a narrow reading listed for the author. Modeled by Claude Code for the bond indices task; the author has not reviewed it yet.

## Source

> **[D-9]** M0 data paths: XTB through uploaded xStation export files; treasury bonds through manual lot entry (series, quantity, purchase date) with valuation computed; an import of the bond service's own export is added if one exists (OQ-22).

docs/DESIGN.md#D-9
