---
title: The price check threshold must be a setting of the core, a fraction greater than 0 and less than 1, and must be 0.01 when it is not set.
status: open
text_ref: docs/DESIGN.md#D-22
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-22
  - lane:strict
semantic_text: The price check threshold must be a setting of the core, a fraction greater than 0 and less than 1, and must be 0.01 when it is not set.
logic_claims:
  - CLAIM-2F4F7A43DE903922
semantic_clauses:
  - The price check threshold must be a setting of the core, a fraction greater than 0 and less than 1, and must be 0.01 when it is not set.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 628acfc9c73759c1986ac25516e77d258bc3decc92ca25e52ba26239c8c742f4
semantic_inventory:
  - claim_key: CLAIM-2F4F7A43DE903922
    claim_text: The price check threshold must be a setting of the core, a fraction greater than 0 and less than 1, and must be 0.01 when it is not set
    role: normative
    status: modeled
    span:
      start: 0
      end: 135
origin:
  kind: agent
  recorded_at: '2026-10-29T05:37:34.159Z'
id: REQ-valuation-price-check-threshold
type: req
---
The price check threshold must be a setting of the core, a fraction greater than 0 and less than 1, and must be 0.01 when it is not set.

## Context

No decision fixes how large a discrepancy is worth flagging, so it is a setting with a default the author can change (listed in docs/implementation-questions.md). One percent is above the few basis points that separate an aggregated close from an official one on a liquid listing, and far below the errors the cross-check exists to catch: a wrong listing or currency, pence read as pounds, a missed split. Refines REQ-AUTO-7374F67D85269872. Modeled by Claude Code for the D-22 cross-check task (impl/cross-check); the author has not reviewed it yet.

## Source

> Prices come from a market-data connector plugin where the user brings the API key (D-51); IBKR Flex prices serve as a cross-check.

docs/DESIGN.md D-22

> Its closes are aggregated rather than each exchange's official close, which is one more reason for D-22's cross-check.

docs/DESIGN.md D-51
