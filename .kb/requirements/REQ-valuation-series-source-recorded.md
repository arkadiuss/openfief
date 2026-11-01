---
title: A series' source must be recorded when it is first chosen and kept until the owner changes it, so a new source ranking changes no series' source and revalues nothing.
status: open
text_ref: docs/DESIGN.md#D-94
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-94
  - lane:strict
semantic_text: A series' source must be recorded when it is first chosen and kept until the owner changes it, so a new source ranking changes no series' source and revalues nothing.
logic_claims:
  - CLAIM-D13D689043B79991
semantic_clauses:
  - A series' source must be recorded when it is first chosen and kept until the owner changes it, so a new source ranking changes no series' source and revalues nothing.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2394657beae19d075289c2c892016f62d92bf888a84e4238fdbbcf31be0df2c3
semantic_inventory:
  - claim_key: CLAIM-D13D689043B79991
    claim_text: A series' source must be recorded when it is first chosen and kept until the owner changes it, so a new source ranking changes no series' source and revalues nothing
    role: normative
    status: modeled
    span:
      start: 0
      end: 165
origin:
  kind: agent
  recorded_at: '2026-11-01T08:55:19.915Z'
id: REQ-valuation-series-source-recorded
type: req
---
A series' source must be recorded when it is first chosen and kept until the owner changes it, so a new source ranking changes no series' source and revalues nothing.

## Context

D-94 as rewritten chooses a series' source from the ranking only when the series first needs a value. Recording the choice is how a series keeps its source; the earlier rule, that a change of ranking revalues every account, no longer holds, because no stored value changes. Modeled by Claude Code for the one-source package; the author has not reviewed it.

## Why this changed

D-94 was rewritten on 2026-11-01: the ranking now only chooses a series' source on first need, so a change of ranking no longer changes any stored value and no longer revalues every account, which REQ-valuation-ranking-revalues required.

## Source

> **[D-94]** Each instrument, currency pair and index has exactly one source of its values: a connector instance, chosen from the owner's ranking of sources for that kind of data when it first needs a value, which the owner can change for that one instrument. A day without a value takes the latest earlier one, and a gap longer than a threshold the owner can set raises an attention item (D-130); a value never comes from another source on the days the chosen one lacks. Every other source's values are stored as cross-checks (D-22). By default the first-party connector of each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

docs/DESIGN.md#D-94
