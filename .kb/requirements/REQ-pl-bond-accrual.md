---
title: 'A bond''s value on a date must follow its issue letter: interest within a period accrues as the yearly rate times the period''s share of a year times days elapsed over the period''s actual days, capitalized types compound completed periods without rounding between them, coupon types count only the current period''s interest, OTS earns its fixed interest only at maturity, and the value is rounded half up to grosze once per bond before multiplying by the lot''s quantity.'
status: open
text_ref: docs/DESIGN.md#D-9
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-9
  - lane:strict
semantic_text: 'A bond''s value on a date must follow its issue letter: interest within a period accrues as the yearly rate times the period''s share of a year times days elapsed over the period''s actual days, capitalized types compound completed periods without rounding between them, coupon types count only the current period''s interest, OTS earns its fixed interest only at maturity, and the value is rounded half up to grosze once per bond before multiplying by the lot''s quantity.'
logic_claims:
  - CLAIM-62D618BFC26C3B50
semantic_clauses:
  - 'A bond''s value on a date must follow its issue letter: interest within a period accrues as the yearly rate times the period''s share of a year times days elapsed over the period''s actual days, capitalized types compound completed periods without rounding between them, coupon types count only the current period''s interest, OTS earns its fixed interest only at maturity, and the value is rounded half up to grosze once per bond before multiplying by the lot''s quantity.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1f32961889133051aad71bbd6ce00866cf5c7f74de1cc9c87326d198227e7720
semantic_inventory:
  - claim_key: CLAIM-62D618BFC26C3B50
    claim_text: 'A bond''s value on a date must follow its issue letter: interest within a period accrues as the yearly rate times the period''s share of a year times days elapsed over the period''s actual days, capitalized types compound completed periods without rounding between them, coupon types count only the current period''s interest, OTS earns its fixed interest only at maturity, and the value is rounded half up to grosze once per bond before multiplying by the lot''s quantity'
    role: normative
    status: modeled
    span:
      start: 0
      end: 467
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:49.632Z'
id: REQ-pl-bond-accrual
type: req
---
A bond's value on a date must follow its issue letter: interest within a period accrues as the yearly rate times the period's share of a year times days elapsed over the period's actual days, capitalized types compound completed periods without rounding between them, coupon types count only the current period's interest, OTS earns its fixed interest only at maturity, and the value is rounded half up to grosze once per bond before multiplying by the lot's quantity.

## Context

Treasury bond valuation is computed, not imported (D-9). The formula is the issue letters' early-redemption value WP_k = N x (1 + r_1) x ... x (1 + r_k x a_k / ACT) - b, rounded to two decimals, with a_k counted from the period's first day inclusive to the valuation day exclusive and ACT the period's actual days. The official interest tables match it on every day checked, including TOS, whose letters word the yearly base as rounded to grosze: the tables do not round it, and this requirement follows the tables (an open point for the author). OTS pays a fixed amount at maturity and nothing on early redemption. Refines REQ-AUTO-30FD669113CEFE3B. Modeled by Claude Code; the author has not reviewed it yet. Checked on 2026-10-28 against the October 2026 issue letters (listy emisyjne nr 94/2026 to 101/2026: OTS0127, ROR1027, DOR1028, TOS1029, COI1030, EDO1036, ROS1032, ROD1038), older letters (TOS0128, TOS0227, COI0128, EDO1133, ROD0236, ROS0228), the interest tables and the worked examples on obligacjeskarbowe.pl.

## Source

> W - wysokość wierzytelności od jednej obligacji w dniu wykupu, w złotych, zaokrąglona do dwóch miejsc po przecinku

EDO1036 issue letter, Załączniki nr 2 and 3; https://www.obligacjeskarbowe.pl/oferta-obligacji/ (each offer page links its issue letter) and https://www.obligacjeskarbowe.pl/listy-emisyjne/; interest tables https://www.obligacjeskarbowe.pl/tabela-odsetkowa/
