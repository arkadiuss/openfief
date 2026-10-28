---
title: The net-if-redeemed value must deduct the series' early-redemption fee per bond, capped at the bond's accrued interest except for coupon types after their first interest period, with no fee at or after maturity, none inside IKE or IKZE, and none for OTS, which instead earns no interest when redeemed early.
status: open
text_ref: docs/DESIGN.md#D-25
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-25
  - lane:strict
semantic_text: The net-if-redeemed value must deduct the series' early-redemption fee per bond, capped at the bond's accrued interest except for coupon types after their first interest period, with no fee at or after maturity, none inside IKE or IKZE, and none for OTS, which instead earns no interest when redeemed early.
logic_claims:
  - CLAIM-4B0EDD6B0BB6A79B
semantic_clauses:
  - The net-if-redeemed value must deduct the series' early-redemption fee per bond, capped at the bond's accrued interest except for coupon types after their first interest period, with no fee at or after maturity, none inside IKE or IKZE, and none for OTS, which instead earns no interest when redeemed early.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 73e33c857e4e19fcc27744e4900b8505a2158fcdb6ddaf36e0204b4d13fb7d36
semantic_inventory:
  - claim_key: CLAIM-4B0EDD6B0BB6A79B
    claim_text: The net-if-redeemed value must deduct the series' early-redemption fee per bond, capped at the bond's accrued interest except for coupon types after their first interest period, with no fee at or after maturity, none inside IKE or IKZE, and none for OTS, which instead earns no interest when redeemed early
    role: exception
    status: modeled
    span:
      start: 0
      end: 306
origin:
  kind: agent
  recorded_at: '2026-10-28T05:22:01.244Z'
id: REQ-pl-early-redemption-fee-cap
type: req
---
The net-if-redeemed value must deduct the series' early-redemption fee per bond, capped at the bond's accrued interest except for coupon types after their first interest period, with no fee at or after maturity, none inside IKE or IKZE, and none for OTS, which instead earns no interest when redeemed early.

## Context

D-25 makes the default bond value net of the early-redemption fee. The fee is series data. Capitalized types: 'dla WP_k < 100 WP_k = 100'. ROR, DOR and COI: capped 'w pierwszym okresie odsetkowym', then the full amount 'począwszy od drugiego okresu odsetkowego'. The letters of ROR, DOR, TOS, COI and EDO waive it for 'wypłaty lub wypłaty transferowej z IKE lub IKZE'; valuation assumes a withdrawal, as it does for the IKE tax exemption. Refines REQ-AUTO-E10BE8CE13E67657. Modeled by Claude Code; the author has not reviewed it yet. Checked on 2026-10-28 against the October 2026 issue letters (listy emisyjne nr 94/2026 to 101/2026: OTS0127, ROR1027, DOR1028, TOS1029, COI1030, EDO1036, ROS1032, ROD1038), older letters (TOS0128, TOS0227, COI0128, EDO1133, ROD0236, ROS0228), the interest tables and the worked examples on obligacjeskarbowe.pl.

## Source

> należność wypłacona z tytułu przedterminowego wykupu jednej obligacji jest pomniejszana o kwotę narosłych odsetek, ale nie wyższą niż 3,00 zł (trzy złote); potrącenia nie dokonuje się wobec posiadaczy obligacji dokonujących wypłaty lub wypłaty transferowej z IKE lub IKZE

EDO1036 issue letter, ust. 26 pkt 5; https://www.obligacjeskarbowe.pl/oferta-obligacji/ (each offer page links its issue letter) and https://www.obligacjeskarbowe.pl/listy-emisyjne/

> the default is net-if-redeemed-today (after the early-redemption fee and Belka tax, no tax inside IKE)

docs/DESIGN.md D-25
