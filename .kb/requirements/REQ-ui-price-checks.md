---
title: The built-in UI must list every flagged price check on its own page, with both prices, their sources, the difference in percent and the threshold, and must count and link them on the overview next to the flagged balance assertions, for any session, without changing any price.
status: open
text_ref: docs/DESIGN.md#D-22
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-22
  - lane:strict
semantic_text: The built-in UI must list every flagged price check on its own page, with both prices, their sources, the difference in percent and the threshold, and must count and link them on the overview next to the flagged balance assertions, for any session, without changing any price.
logic_claims:
  - CLAIM-E2A84E76E83568C6
semantic_clauses:
  - The built-in UI must list every flagged price check on its own page, with both prices, their sources, the difference in percent and the threshold, and must count and link them on the overview next to the flagged balance assertions, for any session, without changing any price.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9cecdcdbf02d2e44d3d39d5e92df1b8f5317dfe5afd91eab209f63c092042009
semantic_inventory:
  - claim_key: CLAIM-E2A84E76E83568C6
    claim_text: The built-in UI must list every flagged price check on its own page, with both prices, their sources, the difference in percent and the threshold, and must count and link them on the overview next to the flagged balance assertions, for any session, without changing any price
    role: normative
    status: modeled
    span:
      start: 0
      end: 275
origin:
  kind: agent
  recorded_at: '2026-10-29T07:03:15.687Z'
id: REQ-ui-price-checks
type: req
---
The built-in UI must list every flagged price check on its own page, with both prices, their sources, the difference in percent and the threshold, and must count and link them on the overview next to the flagged balance assertions, for any session, without changing any price.

## Context

D-22 makes IBKR's prices a cross-check; a flagged check is for the owner to look at, as a flagged balance assertion is (D-18), so the UI shows them the same way. Where they appear is the UI's reading; no decision places them. Refines REQ-AUTO-7374F67D85269872. Modeled by Claude Code for the D-22 cross-check task (impl/cross-check); the author has not reviewed it yet.

## Source

> Prices come from a market-data connector plugin where the user brings the API key (D-51); IBKR Flex prices serve as a cross-check.

docs/DESIGN.md D-22
