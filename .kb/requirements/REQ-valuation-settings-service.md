---
title: 'The core must serve the owner''s settings to view plugins as SettingsService: the reporting currency, the number format with whether it is the default, the formats on offer, the tax wrappers and the marginal rate, and changes of the number format to an offered one and of the marginal rate, refusing anything else with its code.'
status: open
text_ref: docs/DESIGN.md#D-79
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-79
  - lane:strict
semantic_text: 'The core must serve the owner''s settings to view plugins as SettingsService: the reporting currency, the number format with whether it is the default, the formats on offer, the tax wrappers and the marginal rate, and changes of the number format to an offered one and of the marginal rate, refusing anything else with its code.'
logic_claims:
  - CLAIM-4A20072436B6B23A
semantic_clauses:
  - 'The core must serve the owner''s settings to view plugins as SettingsService: the reporting currency, the number format with whether it is the default, the formats on offer, the tax wrappers and the marginal rate, and changes of the number format to an offered one and of the marginal rate, refusing anything else with its code.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9df71fda42fad4e7bffc9f1ed3c93d81caa0092c4a474217f6cad8e7489d8c7a
semantic_inventory:
  - claim_key: CLAIM-4A20072436B6B23A
    claim_text: 'The core must serve the owner''s settings to view plugins as SettingsService: the reporting currency, the number format with whether it is the default, the formats on offer, the tax wrappers and the marginal rate, and changes of the number format to an offered one and of the marginal rate, refusing anything else with its code'
    role: normative
    status: modeled
    span:
      start: 0
      end: 326
origin:
  kind: agent
  recorded_at: '2026-10-29T22:59:17.529Z'
id: REQ-valuation-settings-service
type: req
---
The core must serve the owner's settings to view plugins as SettingsService: the reporting currency, the number format with whether it is the default, the formats on offer, the tax wrappers and the marginal rate, and changes of the number format to an offered one and of the marginal rate, refusing anything else with its code.

## Context

D-79 makes each kind of owner entry its own owner-only service; D-113's number format and D-78's marginal rate are the owner's settings. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Source

> **[D-79]** What the owner enters by hand reaches the core through one service per kind of entry: bond lots, the model portfolio, accounts, statement uploads and any later kind. Each is owner-only, versioned on its own, and called only by a view plugin acting for an owner session (D-69).

docs/DESIGN.md#D-79
