---
title: The Polish jurisdiction pack's default number format must group thousands with a space and use a decimal comma, as in 5 000,00 PLN.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-113-2
  - lane:strict
text_ref: design-decisions:D-113#2
origin:
  kind: import
  ref: design-decisions:D-113#2
  recorded_at: '2026-10-29T11:40:00.000Z'
semantic_text: The Polish jurisdiction pack's default number format must group thousands with a space and use a decimal comma, as in 5 000,00 PLN.
logic_claims:
  - CLAIM-53DFCD435517E3E0
semantic_clauses:
  - The Polish jurisdiction pack's default number format must group thousands with a space and use a decimal comma, as in 5 000,00 PLN.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 67a0cc197f77231d01fcc319bdd868989481bb5c4b29e394d058805e46d4e595
semantic_inventory:
  - claim_key: CLAIM-53DFCD435517E3E0
    claim_text: The Polish jurisdiction pack's default number format must group thousands with a space and use a decimal comma, as in 5 000,00 PLN
    role: normative
    status: modeled
    span:
      start: 0
      end: 130
id: REQ-jurisdiction-pl-number-format
type: req
---
The Polish jurisdiction pack's default number format must group thousands with a space and use a decimal comma, as in 5 000,00 PLN.

## Context

D-113 makes the number format a setting whose default comes from the jurisdiction pack, so the Polish pack must say what Polish readers expect: the design system's example, `5 000,00 PLN`. The author chose the setting with a pack default when the implementation asked which locale the UI writes amounts in.

## Source

> The built-in UI writes amounts in a number format: a setting next to the reporting currency, whose default comes from the jurisdiction pack. The Polish pack's default groups thousands with a space and uses a decimal comma, as in `5 000,00 PLN`. The format changes only the separators, never the digits or the currency, and the charts' axes and tooltips use it too.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-113#2
