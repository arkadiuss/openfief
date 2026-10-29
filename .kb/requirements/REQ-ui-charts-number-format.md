---
title: The built-in UI's chart axes and tooltips must write amounts in the owner's number format.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-113-4
  - lane:strict
text_ref: design-decisions:D-113#4
origin:
  kind: import
  ref: design-decisions:D-113#4
  recorded_at: '2026-10-29T11:40:00.000Z'
semantic_text: The built-in UI's chart axes and tooltips must write amounts in the owner's number format.
logic_claims:
  - CLAIM-69D0EA964920DF33
semantic_clauses:
  - The built-in UI's chart axes and tooltips must write amounts in the owner's number format.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 62ced7e38787fb12c35b3399849a3e0860740445808f2412b82c314c4234474f
semantic_inventory:
  - claim_key: CLAIM-69D0EA964920DF33
    claim_text: The built-in UI's chart axes and tooltips must write amounts in the owner's number format
    role: normative
    status: modeled
    span:
      start: 0
      end: 89
id: REQ-ui-charts-number-format
type: req
---
The built-in UI's chart axes and tooltips must write amounts in the owner's number format.

## Context

The UI's charts used Chart.js's own number format on their axes, so a page could write the same amount two ways. D-113 makes the axes and tooltips follow the owner's number format like the rest of the UI. The author recorded it with the number-format setting.

## Source

> The built-in UI writes amounts in a number format: a setting next to the reporting currency, whose default comes from the jurisdiction pack. The Polish pack's default groups thousands with a space and uses a decimal comma, as in `5 000,00 PLN`. The format changes only the separators, never the digits or the currency, and the charts' axes and tooltips use it too.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-113#4
