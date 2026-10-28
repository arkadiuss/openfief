---
title: Polish tax must convert foreign income at the NBP table A rate of the last business day before the income date, never the rate of the income date itself.
status: open
text_ref: docs/DESIGN.md#D-20
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-20
  - lane:strict
semantic_text: Polish tax must convert foreign income at the NBP table A rate of the last business day before the income date, never the rate of the income date itself.
logic_claims:
  - CLAIM-A3B8CE8965078D19
semantic_clauses:
  - Polish tax must convert foreign income at the NBP table A rate of the last business day before the income date, never the rate of the income date itself.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a84c81fb996b6d0afb3d1bb731fa3db41564a4cde98ad223a005c55eb85c01b3
semantic_inventory:
  - claim_key: CLAIM-A3B8CE8965078D19
    claim_text: Polish tax must convert foreign income at the NBP table A rate of the last business day before the income date, never the rate of the income date itself
    role: normative
    status: modeled
    span:
      start: 0
      end: 152
origin:
  kind: agent
  recorded_at: '2026-10-28T05:22:23.202Z'
id: REQ-pl-tax-fx-rate-date
type: req
---
Polish tax must convert foreign income at the NBP table A rate of the last business day before the income date, never the rate of the income date itself.

## Context

D-20 says the Polish jurisdiction pack always uses NBP rates for tax. Art. 11a ust. 1 of the PIT act takes the NBP average rate of the last business day before the day the income is received. NBP publishes only on business days, so the latest table dated on or before the day before the event is that rate. Refines REQ-AUTO-C17116F5350E97D6. Modeled by Claude Code for the M0 valuation task; the author has not reviewed it yet.

## Source

> The Polish jurisdiction pack always uses NBP rates for tax, whatever the UI uses.

docs/DESIGN.md D-20
