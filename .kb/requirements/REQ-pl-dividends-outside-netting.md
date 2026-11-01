---
title: Dividends must stay out of the netting of gains and losses.
status: open
text_ref: design-decisions:D-150#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-150-2
  - lane:strict
semantic_text: Dividends must stay out of the netting of gains and losses.
logic_claims:
  - CLAIM-59C51EA844529ABC
semantic_clauses:
  - Dividends must stay out of the netting of gains and losses.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 04fff68d12a5a1bc9eec905628c8d690a1d5d0030f62c1774b1d87448853200d
semantic_inventory:
  - claim_key: CLAIM-59C51EA844529ABC
    claim_text: Dividends must stay out of the netting of gains and losses
    role: normative
    status: modeled
    span:
      start: 0
      end: 58
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-150#2
  recorded_at: '2026-11-01T12:28:17.064Z'
id: REQ-pl-dividends-outside-netting
type: req
---
Dividends must stay out of the netting of gains and losses.

## Context

The author recorded it on 2026-11-01, answering the implementation's question #45. The decision gives the reason: Polish tax counts dividends as a separate source of income that capital losses don't offset. D-128 offsets the tax year's gains and losses (REQ-pl-totals-tax-pit38). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The tax cost of a sale includes its buy fees, as PIT-38 deducts them, in each row's tax and in the totals alike. Dividends stay out of the netting, because Polish tax counts them as a separate source of income that capital losses don't offset.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-150#2
