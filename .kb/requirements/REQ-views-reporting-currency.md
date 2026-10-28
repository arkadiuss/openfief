---
title: The read API must report totals, valuations and the allocation in the configured reporting currency next to each value's original currency, and must refuse to start with a reporting currency other than PLN with a typed unsupported reporting currency error until cross rates are decided.
status: open
text_ref: docs/DESIGN.md#D-62
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-62
  - lane:strict
semantic_text: The read API must report totals, valuations and the allocation in the configured reporting currency next to each value's original currency, and must refuse to start with a reporting currency other than PLN with a typed unsupported reporting currency error until cross rates are decided.
logic_claims:
  - CLAIM-65B29AB56E609541
semantic_clauses:
  - The read API must report totals, valuations and the allocation in the configured reporting currency next to each value's original currency, and must refuse to start with a reporting currency other than PLN with a typed unsupported reporting currency error until cross rates are decided.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 47f2fe9cc7f098adb51b4003817223e59877a6b10a3ac80401c343b4c6157837
semantic_inventory:
  - claim_key: CLAIM-65B29AB56E609541
    claim_text: The read API must report totals, valuations and the allocation in the configured reporting currency next to each value's original currency, and must refuse to start with a reporting currency other than PLN with a typed unsupported reporting currency error until cross rates are decided
    role: normative
    status: modeled
    span:
      start: 0
      end: 285
origin:
  kind: agent
  recorded_at: '2026-10-28T08:06:29.326Z'
id: REQ-views-reporting-currency
type: req
---
The read API must report totals, valuations and the allocation in the configured reporting currency next to each value's original currency, and must refuse to start with a reporting currency other than PLN with a typed unsupported reporting currency error until cross rates are decided.

## Context

D-62 makes the reporting currency a setting (default PLN from the Polish pack) and says no code outside the pack assumes PLN, so the views' pure rules take it as a parameter. T-12 is open: NBP table A quotes only against PLN, so the orchestrator asked to support PLN reporting only and return a typed error otherwise. Modeled by Claude Code for the M0 read API task; the author has not reviewed it yet.

## Source

> Totals, charts and the model comparison are shown in the reporting currency: a setting whose default comes from the jurisdiction pack, PLN for the Polish pack. Apart from the Polish pack's tax rules, no code assumes PLN.

docs/DESIGN.md D-62

> D-62 makes the reporting currency a setting, but D-20 converts amounts with NBP table A, which quotes every currency only against PLN. A reporting currency other than PLN needs cross rates through PLN or another FX plugin.

docs/DESIGN.md T-12
