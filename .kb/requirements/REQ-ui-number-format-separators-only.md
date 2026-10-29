---
title: The number format must change only an amount's separators, never its digits or its currency.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-113-3
  - lane:strict
text_ref: design-decisions:D-113#3
origin:
  kind: import
  ref: design-decisions:D-113#3
  recorded_at: '2026-10-29T11:40:00.000Z'
semantic_text: The number format must change only an amount's separators, never its digits or its currency.
logic_claims:
  - CLAIM-C5056510FDD4C018
semantic_clauses:
  - The number format must change only an amount's separators, never its digits or its currency.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: cf2cd6b3e6eae777c2ab661e7bf3e15ee4874f84a62b4829c1b0f611f5975cd1
semantic_inventory:
  - claim_key: CLAIM-C5056510FDD4C018
    claim_text: The number format must change only an amount's separators, never its digits or its currency
    role: normative
    status: modeled
    span:
      start: 0
      end: 91
id: REQ-ui-number-format-separators-only
type: req
---
The number format must change only an amount's separators, never its digits or its currency.

## Context

Every amount keeps its original currency (D-20) and the UI writes the read API's exact decimal strings, so a number format may only change how digits are grouped and which mark separates the decimals. The author recorded this with the number-format setting in D-113.

## Source

> The built-in UI writes amounts in a number format: a setting next to the reporting currency, whose default comes from the jurisdiction pack. The Polish pack's default groups thousands with a space and uses a decimal comma, as in `5 000,00 PLN`. The format changes only the separators, never the digits or the currency, and the charts' axes and tooltips use it too.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-113#3
