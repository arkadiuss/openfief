---
title: The read API must give any reader the reporting currency, the number format with whether it is the pack's default, and the jurisdiction pack's tax wrappers, the default first.
status: open
text_ref: docs/DESIGN.md#D-113
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-113
  - lane:strict
semantic_text: The read API must give any reader the reporting currency, the number format with whether it is the pack's default, and the jurisdiction pack's tax wrappers, the default first.
logic_claims:
  - CLAIM-772AA50991BB1799
semantic_clauses:
  - The read API must give any reader the reporting currency, the number format with whether it is the pack's default, and the jurisdiction pack's tax wrappers, the default first.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a9b7d8a63c75e4f23d3a5bea008e0822df4466dbdaf18b7d9a4fff6775e71ef9
semantic_inventory:
  - claim_key: CLAIM-772AA50991BB1799
    claim_text: The read API must give any reader the reporting currency, the number format with whether it is the pack's default, and the jurisdiction pack's tax wrappers, the default first
    role: normative
    status: modeled
    span:
      start: 0
      end: 174
origin:
  kind: agent
  recorded_at: '2026-10-30T00:05:28.191Z'
id: REQ-views-settings
type: req
---
The read API must give any reader the reporting currency, the number format with whether it is the pack's default, and the jurisdiction pack's tax wrappers, the default first.

## Context

D-113 has the UI write amounts in the owner's format, for any session; D-62 and D-76 name the currency and the wrappers. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Source

> **[D-113]** The built-in UI writes amounts in a number format: a setting next to the reporting currency, whose default comes from the jurisdiction pack. The Polish pack's default groups thousands with a space and uses a decimal comma, as in `5 000,00 PLN`. The format changes only the separators, never the digits or the currency, and the charts' axes and tooltips use it too.

docs/DESIGN.md#D-113
