---
title: A number format must be a group separator (space, comma, full stop, apostrophe or none) and a decimal separator (comma or full stop), different from each other, stored by a key naming both, such as space-comma.
status: open
text_ref: docs/DESIGN.md#D-113
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-113
  - lane:strict
semantic_text: A number format must be a group separator (space, comma, full stop, apostrophe or none) and a decimal separator (comma or full stop), different from each other, stored by a key naming both, such as space-comma.
logic_claims:
  - CLAIM-53A571BD60D461B7
semantic_clauses:
  - A number format must be a group separator (space, comma, full stop, apostrophe or none) and a decimal separator (comma or full stop), different from each other, stored by a key naming both, such as space-comma.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fd7afb10b008d80d856f4d2b7488b9b218b7b4a043faa9ccda711b7fbecbe2ac
semantic_inventory:
  - claim_key: CLAIM-53A571BD60D461B7
    claim_text: A number format must be a group separator (space, comma, full stop, apostrophe or none) and a decimal separator (comma or full stop), different from each other, stored by a key naming both, such as space-comma
    role: normative
    status: modeled
    span:
      start: 0
      end: 209
origin:
  kind: agent
  recorded_at: '2026-10-29T21:48:45.270Z'
id: REQ-jurisdiction-number-format-choices
type: req
---
A number format must be a group separator (space, comma, full stop, apostrophe or none) and a decimal separator (comma or full stop), different from each other, stored by a key naming both, such as space-comma.

## Context

D-113 makes the format change only the separators; the list of choices is this package's. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Source

> **[D-113]** The built-in UI writes amounts in a number format: a setting next to the reporting currency, whose default comes from the jurisdiction pack. The Polish pack's default groups thousands with a space and uses a decimal comma, as in `5 000,00 PLN`. The format changes only the separators, never the digits or the currency, and the charts' axes and tooltips use it too.

docs/DESIGN.md#D-113
