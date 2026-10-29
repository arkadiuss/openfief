---
title: The built-in UI must write amounts in the owner's number format, a setting next to the reporting currency whose default comes from the jurisdiction pack.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-113-1
  - lane:strict
text_ref: design-decisions:D-113#1
origin:
  kind: import
  ref: design-decisions:D-113#1
  recorded_at: '2026-10-29T11:40:00.000Z'
semantic_text: The built-in UI must write amounts in the owner's number format, a setting next to the reporting currency whose default comes from the jurisdiction pack.
logic_claims:
  - CLAIM-7FA32B20135A302A
semantic_clauses:
  - The built-in UI must write amounts in the owner's number format, a setting next to the reporting currency whose default comes from the jurisdiction pack.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 1507d12e6193d1890e468506e2350c33273d51b0fa62d4e408209110ae0d1aec
semantic_inventory:
  - claim_key: CLAIM-7FA32B20135A302A
    claim_text: The built-in UI must write amounts in the owner's number format, a setting next to the reporting currency whose default comes from the jurisdiction pack
    role: normative
    status: modeled
    span:
      start: 0
      end: 152
id: REQ-ui-number-format-setting
type: req
---
The built-in UI must write amounts in the owner's number format, a setting next to the reporting currency whose default comes from the jurisdiction pack.

## Context

The design system asked for amounts formatted for the reader's locale, but no decision said which locale the UI uses: its pages are in English and OpenFief is built for Polish investors first. The implementation raised it, and the author chose a setting next to the reporting currency (D-62) with the jurisdiction pack's default, recorded as D-113.

## Source

> The built-in UI writes amounts in a number format: a setting next to the reporting currency, whose default comes from the jurisdiction pack. The Polish pack's default groups thousands with a space and uses a decimal comma, as in `5 000,00 PLN`. The format changes only the separators, never the digits or the currency, and the charts' axes and tooltips use it too.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-113#1
