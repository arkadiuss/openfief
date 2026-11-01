---
title: The rules version the stored values are computed under must include the version of the rules by which a series takes its values from its sources, so a change of those rules revalues every account at start-up.
status: open
text_ref: docs/DESIGN.md#D-94
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-94
  - lane:strict
semantic_text: The rules version the stored values are computed under must include the version of the rules by which a series takes its values from its sources, so a change of those rules revalues every account at start-up.
logic_claims:
  - CLAIM-E846C552511D8FF7
semantic_clauses:
  - The rules version the stored values are computed under must include the version of the rules by which a series takes its values from its sources, so a change of those rules revalues every account at start-up.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3523b7ce3c6109fa17fd0b13a3ee3bf4ee88de902d605f3979ab3624d5c3159e
semantic_inventory:
  - claim_key: CLAIM-E846C552511D8FF7
    claim_text: The rules version the stored values are computed under must include the version of the rules by which a series takes its values from its sources, so a change of those rules revalues every account at start-up
    role: normative
    status: modeled
    span:
      start: 0
      end: 207
origin:
  kind: agent
  recorded_at: '2026-11-01T08:56:08.579Z'
id: REQ-valuation-source-rules-version
type: req
---
The rules version the stored values are computed under must include the version of the rules by which a series takes its values from its sources, so a change of those rules revalues every account at start-up.

## Context

Data stored before D-94 was rewritten was valued with values of several sources per series. The migration records no source; series get theirs on first need, and the new rules version makes that need happen for every account at once, so revaluation gives consistent results. Built by Claude Code; not reviewed by the author.

## Source

> **[D-94]** Each instrument, currency pair and index has exactly one source of its values: a connector instance, chosen from the owner's ranking of sources for that kind of data when it first needs a value, which the owner can change for that one instrument. A day without a value takes the latest earlier one, and a gap longer than a threshold the owner can set raises an attention item (D-130); a value never comes from another source on the days the chosen one lacks. Every other source's values are stored as cross-checks (D-22). By default the first-party connector of each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

docs/DESIGN.md#D-94
