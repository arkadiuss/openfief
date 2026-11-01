---
title: The owner must be able to classify an instrument in one of the pack's exemption classes through the SettingsService, stored with the instrument and replacing an earlier choice, without revaluing.
status: open
text_ref: docs/DESIGN.md#D-132
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-132
  - lane:strict
semantic_text: The owner must be able to classify an instrument in one of the pack's exemption classes through the SettingsService, stored with the instrument and replacing an earlier choice, without revaluing.
logic_claims:
  - CLAIM-7695703E7ACF4BD4
semantic_clauses:
  - The owner must be able to classify an instrument in one of the pack's exemption classes through the SettingsService, stored with the instrument and replacing an earlier choice, without revaluing.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 81aa20432f4e3632df65da5cc7c31f446fa15fa68e93efa91ecf758d46bf0f2c
semantic_inventory:
  - claim_key: CLAIM-7695703E7ACF4BD4
    claim_text: The owner must be able to classify an instrument in one of the pack's exemption classes through the SettingsService, stored with the instrument and replacing an earlier choice, without revaluing
    role: normative
    status: modeled
    span:
      start: 0
      end: 194
origin:
  kind: agent
  recorded_at: '2026-10-31T18:34:40.485Z'
id: REQ-valuation-owner-exemption-class
type: req
---
The owner must be able to classify an instrument in one of the pack's exemption classes through the SettingsService, stored with the instrument and replacing an earlier choice, without revaluing.

## Context

D-132: the owner's choice stays with the instrument. The yearly tax is computed when read, so no stored value depends on it. Modeled by Claude Code for the taxes work package; the author has not reviewed it yet.

## Source

> **[D-132]** When the pack can't tell whether an instrument is a savings asset, an investment asset or neither for D-127's exemptions, such as a fund whose policy it can't read, the instrument counts as neither, so the estimate errs high, and an attention item (D-130) asks the owner to classify it once; the owner's choice stays with the instrument.

docs/DESIGN.md#D-132
