---
title: A bond series code must parse into its bond type and maturity month, and a bond lot whose purchase date plus the type's term does not fall in that month must be rejected.
status: open
text_ref: docs/DESIGN.md#D-9
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-9
  - lane:strict
semantic_text: A bond series code must parse into its bond type and maturity month, and a bond lot whose purchase date plus the type's term does not fall in that month must be rejected.
logic_claims:
  - CLAIM-B98560FF594E56D7
semantic_clauses:
  - A bond series code must parse into its bond type and maturity month, and a bond lot whose purchase date plus the type's term does not fall in that month must be rejected.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 53da2e03b6d5e47da7086f4cb10c703f9f33c9b9e52a95a929563d2f9f675a19
semantic_inventory:
  - claim_key: CLAIM-B98560FF594E56D7
    claim_text: A bond series code must parse into its bond type and maturity month, and a bond lot whose purchase date plus the type's term does not fall in that month must be rejected
    role: normative
    status: modeled
    span:
      start: 0
      end: 169
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:41.499Z'
id: REQ-pl-series-code
type: req
---
A bond series code must parse into its bond type and maturity month, and a bond lot whose purchase date plus the type's term does not fall in that month must be rejected.

## Context

Bond lots are entered by hand with series, quantity and purchase date (D-9), so a typo in the series code would silently value the wrong product. The code's three letters give the bond type; its MMYY gives the maturity month, and each bond's term runs from its own purchase date. Every series is sold only from the 1st to the last day of its own month (issue letters ust. 4); a purchase by exchange is dated on the old bond's maturity day, which falls in the month the new series is sold, so the rule holds for exchanges too. Refines REQ-AUTO-FD50BCD71546E83F. Modeled by Claude Code; the author has not reviewed it yet. Checked on 2026-10-28 against the October 2026 issue letters (listy emisyjne nr 94/2026 to 101/2026: OTS0127, ROR1027, DOR1028, TOS1029, COI1030, EDO1036, ROS1032, ROD1038), older letters (TOS0128, TOS0227, COI0128, EDO1133, ROD0236, ROS0228), the interest tables and the worked examples on obligacjeskarbowe.pl.

## Source

> Obligacje są oferowane do sprzedaży w dniach od 1 do 31 października 2026 r.

EDO1036 issue letter, ust. 4; https://www.obligacjeskarbowe.pl/oferta-obligacji/ (each offer page links its issue letter) and https://www.obligacjeskarbowe.pl/listy-emisyjne/

> treasury bonds through manual lot entry (series, quantity, purchase date) with valuation computed

docs/DESIGN.md D-9
