---
title: Totals, charts and the model comparison must be shown in the reporting currency, a setting whose default comes from the jurisdiction pack; apart from the Polish pack's tax rules, no code may assume PLN.
status: open
text_ref: design-decisions:D-62#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-62-1
  - lane:strict
semantic_text: Totals, charts and the model comparison must be shown in the reporting currency, a setting whose default comes from the jurisdiction pack; apart from the Polish pack's tax rules, no code may assume PLN.
logic_claims:
  - CLAIM-E32C311FD529E7D4
semantic_clauses:
  - Totals, charts and the model comparison must be shown in the reporting currency, a setting whose default comes from the jurisdiction pack; apart from the Polish pack's tax rules, no code may assume PLN.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3680bed461268709df0d03a9cb07eab017abf1e681f25a3657dc9cf963d9a8d3
semantic_inventory:
  - claim_key: CLAIM-E32C311FD529E7D4
    claim_text: Totals, charts and the model comparison must be shown in the reporting currency, a setting whose default comes from the jurisdiction pack; apart from the Polish pack's tax rules, no code may assume PLN
    role: exception
    status: modeled
    span:
      start: 0
      end: 201
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-62#1
  recorded_at: '2026-10-28T06:18:42.473Z'
id: REQ-valuation-reporting-currency
type: req
---
Totals, charts and the model comparison must be shown in the reporting currency, a setting whose default comes from the jurisdiction pack; apart from the Polish pack's tax rules, no code may assume PLN.

## Context

Double-entry keeps cash, withholding tax and currency conversion consistent by construction, and it gives a later lease a natural home as a sub-account.

## Source

> Totals, charts and the model comparison are shown in the reporting currency: a setting whose default comes from the jurisdiction pack, PLN for the Polish pack. Apart from the Polish pack's tax rules, no code assumes PLN.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-62#1
