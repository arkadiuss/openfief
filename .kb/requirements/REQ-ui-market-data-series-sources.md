---
title: The built-in UI's settings page must show each kind's gap threshold and each series' source with the sources that have a value of it, and let the owner set a threshold or choose another source through the core's MarketDataService.
status: open
text_ref: docs/DESIGN.md#D-94
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-94
  - lane:strict
semantic_text: The built-in UI's settings page must show each kind's gap threshold and each series' source with the sources that have a value of it, and let the owner set a threshold or choose another source through the core's MarketDataService.
logic_claims:
  - CLAIM-F7FD37CAB919F7E2
semantic_clauses:
  - The built-in UI's settings page must show each kind's gap threshold and each series' source with the sources that have a value of it, and let the owner set a threshold or choose another source through the core's MarketDataService.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: af5ad1aa041d6de87e8576eca527a17dfbf8a4d11df65a09818be98ac2f0e1bf
semantic_inventory:
  - claim_key: CLAIM-F7FD37CAB919F7E2
    claim_text: The built-in UI's settings page must show each kind's gap threshold and each series' source with the sources that have a value of it, and let the owner set a threshold or choose another source through the core's MarketDataService
    role: normative
    status: modeled
    span:
      start: 0
      end: 229
origin:
  kind: agent
  recorded_at: '2026-11-01T08:56:26.246Z'
id: REQ-ui-market-data-series-sources
type: req
---
The built-in UI's settings page must show each kind's gap threshold and each series' source with the sources that have a value of it, and let the owner set a threshold or choose another source through the core's MarketDataService.

## Context

D-94 lets the owner change one instrument's source and set the gap threshold; the built-in UI is the owner's surface for it (D-79). Built by Claude Code to docs/dev/design-system.md; not reviewed by the author.

## Source

> **[D-94]** Each instrument, currency pair and index has exactly one source of its values: a connector instance, chosen from the owner's ranking of sources for that kind of data when it first needs a value, which the owner can change for that one instrument. A day without a value takes the latest earlier one, and a gap longer than a threshold the owner can set raises an attention item (D-130); a value never comes from another source on the days the chosen one lacks. Every other source's values are stored as cross-checks (D-22). By default the first-party connector of each kind ranks first: NBP for exchange rates, EODHD for prices, and the treasury bond connector for CPI and the NBP reference rate.

docs/DESIGN.md#D-94
