---
title: The read API must report, per account, the date it is valued through and whether a revaluation is pending or running.
status: open
text_ref: design-decisions:D-36#6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-36-6
  - lane:strict
semantic_text: The read API must report, per account, the date it is valued through and whether a revaluation is pending or running.
logic_claims:
  - CLAIM-5401F3F03541E1CB
semantic_clauses:
  - The read API must report, per account, the date it is valued through and whether a revaluation is pending or running.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: adf7d4133f4f73236d28ca080e989f7a97d3f2bb376e3e76fc9f46403da02894
semantic_inventory:
  - claim_key: CLAIM-5401F3F03541E1CB
    claim_text: The read API must report, per account, the date it is valued through and whether a revaluation is pending or running
    role: normative
    status: modeled
    span:
      start: 0
      end: 116
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-36#6
  recorded_at: '2026-10-28T06:29:52.370Z'
id: REQ-decision-d36-6-api-valuation-status
type: req
---
The read API must report, per account, the date it is valued through and whether a revaluation is pending or running.

## Context

The author decided D-36 in docs/DESIGN.md section 3.3; this requirement imports its claim D-36#6 as a firm rule, so changing it needs the author. Reason not stated in the decision. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The read API reports, per account, the date it is valued through and whether a revaluation is pending or running, and the UI shows it.

Source: OpenFief author decisions (docs/DESIGN.md section 3.3) - D-36#6
