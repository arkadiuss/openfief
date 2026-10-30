---
title: The owner's settings must give anyone the reporting currency and the number format, the pack's default until the owner chooses one, and let only an owner session read the wrappers and the marginal rate and change the number format and the marginal rate.
status: open
text_ref: docs/DESIGN.md#D-113
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-113
  - lane:strict
semantic_text: The owner's settings must give anyone the reporting currency and the number format, the pack's default until the owner chooses one, and let only an owner session read the wrappers and the marginal rate and change the number format and the marginal rate.
logic_claims:
  - CLAIM-6468F85BFB7DAC9F
semantic_clauses:
  - The owner's settings must give anyone the reporting currency and the number format, the pack's default until the owner chooses one, and let only an owner session read the wrappers and the marginal rate and change the number format and the marginal rate.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 10b7f92362c2cf555055d93000c44c0df4f5fddc8a527300bdab0dc9f37d2e65
semantic_inventory:
  - claim_key: CLAIM-6468F85BFB7DAC9F
    claim_text: The owner's settings must give anyone the reporting currency and the number format, the pack's default until the owner chooses one, and let only an owner session read the wrappers and the marginal rate and change the number format and the marginal rate
    role: normative
    status: modeled
    span:
      start: 0
      end: 252
origin:
  kind: agent
  recorded_at: '2026-10-29T21:49:25.225Z'
id: REQ-valuation-owner-settings
type: req
---
The owner's settings must give anyone the reporting currency and the number format, the pack's default until the owner chooses one, and let only an owner session read the wrappers and the marginal rate and change the number format and the marginal rate.

## Context

D-113 puts the number format next to the reporting currency; D-78 has the owner enter the marginal rate once; D-79 makes each kind of owner entry an owner-only service. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Source

> **[D-113]** The built-in UI writes amounts in a number format: a setting next to the reporting currency, whose default comes from the jurisdiction pack. The Polish pack's default groups thousands with a space and uses a decimal comma, as in `5 000,00 PLN`. The format changes only the separators, never the digits or the currency, and the charts' axes and tooltips use it too.

docs/DESIGN.md#D-113
