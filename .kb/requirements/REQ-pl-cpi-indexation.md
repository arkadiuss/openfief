---
title: From the second interest period, an inflation-indexed bond's rate must be the CPI reading plus the series margin and a ROR or DOR bond's rate the NBP reference rate plus the margin, with a negative reading counted as zero in both cases, and OTS and TOS must keep their first-period rate.
status: open
text_ref: docs/DESIGN.md#D-9
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-9
  - lane:strict
semantic_text: From the second interest period, an inflation-indexed bond's rate must be the CPI reading plus the series margin and a ROR or DOR bond's rate the NBP reference rate plus the margin, with a negative reading counted as zero in both cases, and OTS and TOS must keep their first-period rate.
logic_claims:
  - CLAIM-F99B950926939D27
semantic_clauses:
  - From the second interest period, an inflation-indexed bond's rate must be the CPI reading plus the series margin and a ROR or DOR bond's rate the NBP reference rate plus the margin, with a negative reading counted as zero in both cases, and OTS and TOS must keep their first-period rate.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: cd4bd0503743059c02f2e09ca88530f2619c20656808a41726ad2ae7f9cc71d2
semantic_inventory:
  - claim_key: CLAIM-F99B950926939D27
    claim_text: From the second interest period, an inflation-indexed bond's rate must be the CPI reading plus the series margin and a ROR or DOR bond's rate the NBP reference rate plus the margin, with a negative reading counted as zero in both cases, and OTS and TOS must keep their first-period rate
    role: normative
    status: modeled
    span:
      start: 0
      end: 286
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:54.508Z'
id: REQ-pl-cpi-indexation
type: req
---
From the second interest period, an inflation-indexed bond's rate must be the CPI reading plus the series margin and a ROR or DOR bond's rate the NBP reference rate plus the margin, with a negative reading counted as zero in both cases, and OTS and TOS must keep their first-period rate.

## Context

The bonds connector takes CPI and the NBP reference rate as inputs (DESIGN.md 3 diagram). Readings and margins are data per series and per period. Załącznik nr 1 of every indexed letter gives r = i + m with 'w przypadku gdy i < 0 przyjmuje się że i = 0', for CPI and for the reference rate alike. CPI is the reading GUS announces in the month before the period starts (the month two before it); the reference rate is the one in force on the 10th business day before the first day of the period's calendar month. Refines REQ-AUTO-30FD669113CEFE3B. Modeled by Claude Code; the author has not reviewed it yet. Checked on 2026-10-28 against the October 2026 issue letters (listy emisyjne nr 94/2026 to 101/2026: OTS0127, ROR1027, DOR1028, TOS1029, COI1030, EDO1036, ROS1032, ROD1038), older letters (TOS0128, TOS0227, COI0128, EDO1133, ROD0236, ROS0228), the interest tables and the worked examples on obligacjeskarbowe.pl.

## Source

> r=i+m [...] w przypadku gdy i < 0 przyjmuje się że i = 0

Issue letters of ROR1027, DOR1028, COI1030, EDO1036, ROS1032, ROD1038, Załącznik nr 1; https://www.obligacjeskarbowe.pl/oferta-obligacji/ (each offer page links its issue letter) and https://www.obligacjeskarbowe.pl/listy-emisyjne/
