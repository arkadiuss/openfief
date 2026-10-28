---
title: 'A bond''s value on a date must follow its issue letter: interest within a period accrues as the yearly rate times the period''s share of a year times days elapsed over the period''s actual days, capitalized types compound completed periods, coupon types count only the current period''s interest, and the value is rounded to grosze per bond before multiplying by the lot''s quantity.'
status: open
text_ref: docs/DESIGN.md#D-9
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-9
  - lane:strict
semantic_text: 'A bond''s value on a date must follow its issue letter: interest within a period accrues as the yearly rate times the period''s share of a year times days elapsed over the period''s actual days, capitalized types compound completed periods, coupon types count only the current period''s interest, and the value is rounded to grosze per bond before multiplying by the lot''s quantity.'
logic_claims:
  - CLAIM-7F615AAECB6E52FC
semantic_clauses:
  - 'A bond''s value on a date must follow its issue letter: interest within a period accrues as the yearly rate times the period''s share of a year times days elapsed over the period''s actual days, capitalized types compound completed periods, coupon types count only the current period''s interest, and the value is rounded to grosze per bond before multiplying by the lot''s quantity.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1e29c0bb3ea536b11f4f79552de40b067f673600d49587b822611f47cbe5ffea
semantic_inventory:
  - claim_key: CLAIM-7F615AAECB6E52FC
    claim_text: 'A bond''s value on a date must follow its issue letter: interest within a period accrues as the yearly rate times the period''s share of a year times days elapsed over the period''s actual days, capitalized types compound completed periods, coupon types count only the current period''s interest, and the value is rounded to grosze per bond before multiplying by the lot''s quantity'
    role: normative
    status: modeled
    span:
      start: 0
      end: 377
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:49.632Z'
id: REQ-pl-bond-accrual
type: req
---
A bond's value on a date must follow its issue letter: interest within a period accrues as the yearly rate times the period's share of a year times days elapsed over the period's actual days, capitalized types compound completed periods, coupon types count only the current period's interest, and the value is rounded to grosze per bond before multiplying by the lot's quantity.

## Context

Treasury bond valuation is computed, not imported (D-9), and statutory bond values are one of the things OpenFief must get exactly right (DESIGN.md 3.1). The formula W = N x product(1 + r_n) x (1 + r_k x a/ACT), rounded to one grosz, is the issue letters' early-redemption value of a capitalized bond; coupon types pay completed periods out. The numerator stays exact and is divided once. Refines REQ-AUTO-30FD669113CEFE3B. Modeled by Claude Code for the M0 valuation task; the author has not reviewed it yet. obligacjeskarbowe.pl and gov.pl were not reachable from the agent's sandbox, so the rule was checked only against search-result excerpts of issue letters and secondary calculators.

## Source

> treasury bonds through manual lot entry (series, quantity, purchase date) with valuation computed

docs/DESIGN.md D-9

Issue letters: https://www.obligacjeskarbowe.pl/oferta-obligacji/ and https://www.obligacjeskarbowe.pl/listy-emisyjne/
