---
title: The tax cost of a sale must include its buy fees, in each row's tax and in the totals alike.
status: open
text_ref: design-decisions:D-150#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-150-1
  - lane:strict
semantic_text: The tax cost of a sale must include its buy fees, in each row's tax and in the totals alike.
logic_claims:
  - CLAIM-463BC005B0E9EF54
semantic_clauses:
  - The tax cost of a sale must include its buy fees, in each row's tax and in the totals alike.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4a7808c96df4837b499f78a3e4c1d87bfc4abc1e3411fe85d7922efde928ddf1
semantic_inventory:
  - claim_key: CLAIM-463BC005B0E9EF54
    claim_text: The tax cost of a sale must include its buy fees, in each row's tax and in the totals alike
    role: normative
    status: modeled
    span:
      start: 0
      end: 91
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-150#1
  recorded_at: '2026-11-01T12:28:09.017Z'
id: REQ-pl-sale-tax-cost-includes-buy-fees
type: req
---
The tax cost of a sale must include its buy fees, in each row's tax and in the totals alike.

## Context

The author recorded it on 2026-11-01, answering the implementation's question #45. The decision gives the reason: PIT-38 deducts them. D-128 has a holding carry its own tax and the totals count it as PIT-38 does (REQ-pl-totals-tax-pit38). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The tax cost of a sale includes its buy fees, as PIT-38 deducts them, in each row's tax and in the totals alike. Dividends stay out of the netting, because Polish tax counts them as a separate source of income that capital losses don't offset.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-150#1
