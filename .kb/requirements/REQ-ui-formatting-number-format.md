---
title: The built-in UI must format amounts from the read API's exact decimal strings in the number format its settings give, a space group written as a narrow no-break space, never through binary floating point, give the charts the format's separators for their axes and tooltips, and compute allocation shares with the kernel's decimals, half-even to one decimal place.
status: open
text_ref: docs/DESIGN.md#D-113
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-113
  - lane:strict
semantic_text: The built-in UI must format amounts from the read API's exact decimal strings in the number format its settings give, a space group written as a narrow no-break space, never through binary floating point, give the charts the format's separators for their axes and tooltips, and compute allocation shares with the kernel's decimals, half-even to one decimal place.
logic_claims:
  - CLAIM-A757881093D98DCC
semantic_clauses:
  - The built-in UI must format amounts from the read API's exact decimal strings in the number format its settings give, a space group written as a narrow no-break space, never through binary floating point, give the charts the format's separators for their axes and tooltips, and compute allocation shares with the kernel's decimals, half-even to one decimal place.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 52d5ce4df9043ed8e0e5aa5e756108e39d275ae25235faeb9f317a2a6055a667
semantic_inventory:
  - claim_key: CLAIM-A757881093D98DCC
    claim_text: The built-in UI must format amounts from the read API's exact decimal strings in the number format its settings give, a space group written as a narrow no-break space, never through binary floating point, give the charts the format's separators for their axes and tooltips, and compute allocation shares with the kernel's decimals, half-even to one decimal place
    role: normative
    status: modeled
    span:
      start: 0
      end: 362
origin:
  kind: agent
  recorded_at: '2026-10-30T02:37:41.716Z'
id: REQ-ui-formatting-number-format
type: req
---
The built-in UI must format amounts from the read API's exact decimal strings in the number format its settings give, a space group written as a narrow no-break space, never through binary floating point, give the charts the format's separators for their axes and tooltips, and compute allocation shares with the kernel's decimals, half-even to one decimal place.

## Context

Supersedes REQ-ui-formatting, which grouped amounts with a fixed separator and a decimal point: D-113 writes them in the owner's number format. The narrow no-break space keeps an amount on one line. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Why this changed

It replaces REQ-ui-formatting: D-113 writes amounts in the owner's number format. The superseded requirement is closed.

## Source

> **[D-113]** The built-in UI writes amounts in a number format: a setting next to the reporting currency, whose default comes from the jurisdiction pack. The Polish pack's default groups thousands with a space and uses a decimal comma, as in `5 000,00 PLN`. The format changes only the separators, never the digits or the currency, and the charts' axes and tooltips use it too.

docs/DESIGN.md#D-113
