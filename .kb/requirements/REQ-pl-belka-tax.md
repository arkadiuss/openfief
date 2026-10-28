---
title: The net-if-redeemed value of a taxable bond lot must deduct Belka tax of 19% of accrued interest minus the early-redemption fee plus the discount of a purchase below the nominal, rounded half up to whole grosze, and no tax when that is not positive.
status: open
text_ref: docs/DESIGN.md#D-25
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-25
  - lane:strict
semantic_text: The net-if-redeemed value of a taxable bond lot must deduct Belka tax of 19% of accrued interest minus the early-redemption fee plus the discount of a purchase below the nominal, rounded half up to whole grosze, and no tax when that is not positive.
logic_claims:
  - CLAIM-11F2B03AA90BDEFA
semantic_clauses:
  - The net-if-redeemed value of a taxable bond lot must deduct Belka tax of 19% of accrued interest minus the early-redemption fee plus the discount of a purchase below the nominal, rounded half up to whole grosze, and no tax when that is not positive.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a1e836375b1742b399911bd76b90c1af918dd85f404df08344aab1b333aa6234
semantic_inventory:
  - claim_key: CLAIM-11F2B03AA90BDEFA
    claim_text: The net-if-redeemed value of a taxable bond lot must deduct Belka tax of 19% of accrued interest minus the early-redemption fee plus the discount of a purchase below the nominal, rounded half up to whole grosze, and no tax when that is not positive
    role: normative
    status: modeled
    span:
      start: 0
      end: 248
origin:
  kind: agent
  recorded_at: '2026-10-28T05:22:08.401Z'
id: REQ-pl-belka-tax
type: req
---
The net-if-redeemed value of a taxable bond lot must deduct Belka tax of 19% of accrued interest minus the early-redemption fee plus the discount of a purchase below the nominal, rounded half up to whole grosze, and no tax when that is not positive.

## Context

D-25 makes the default bond value net of Belka tax. Art. 30a ust. 1 pkt 2 of the PIT act taxes 'odsetek i dyskonta od papierów wartościowych' at 19%, and the issuer's FAQ says the discount of a purchase by exchange (99.90 PLN) is taxed. The worked examples deduct the fee before tax (EDO0524: tax on 2.16 = 4.16 - 2). Rounding is open: art. 63 par. 1a of the tax ordinance rounds 'do pełnych groszy w górę' (up), but the issuer's examples round half up (EDO0524 0.41 on 2.16, ROR0623 0.08 on 0.44, COI0423 1.85 on 9.75); the code follows the examples until the author decides. Refines REQ-AUTO-E10BE8CE13E67657. Modeled by Claude Code; the author has not reviewed it yet. Checked on 2026-10-28 against the October 2026 issue letters (listy emisyjne nr 94/2026 to 101/2026: OTS0127, ROR1027, DOR1028, TOS1029, COI1030, EDO1036, ROS1032, ROD1038), older letters (TOS0128, TOS0227, COI0128, EDO1133, ROD0236, ROS0228), the interest tables and the worked examples on obligacjeskarbowe.pl.

## Source

> W przypadku, gdy obligacje były zakupione z dyskontem (np. w drodze zamiany) - opodatkowaniu 19-proc. podatkiem podlega również wartość dyskonta.

https://www.obligacjeskarbowe.pl/faq/

> the default is net-if-redeemed-today (after the early-redemption fee and Belka tax, no tax inside IKE)

docs/DESIGN.md D-25
