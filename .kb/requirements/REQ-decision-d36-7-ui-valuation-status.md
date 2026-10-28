---
title: The UI must show, per account, the date it is valued through and whether a revaluation is pending or running.
status: open
text_ref: design-decisions:D-36#7
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-36-7
  - lane:strict
semantic_text: The UI must show, per account, the date it is valued through and whether a revaluation is pending or running.
logic_claims:
  - CLAIM-AE6DB3324036681F
semantic_clauses:
  - The UI must show, per account, the date it is valued through and whether a revaluation is pending or running.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 10347b15eefe6267b1d712b827d3e9d67b4e162aee3eb50e7ba01397c591d615
semantic_inventory:
  - claim_key: CLAIM-AE6DB3324036681F
    claim_text: The UI must show, per account, the date it is valued through and whether a revaluation is pending or running
    role: normative
    status: modeled
    span:
      start: 0
      end: 108
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-36#7
  recorded_at: '2026-10-28T06:30:01.090Z'
id: REQ-decision-d36-7-ui-valuation-status
type: req
---
The UI must show, per account, the date it is valued through and whether a revaluation is pending or running.

## Context

The author decided D-36 in docs/DESIGN.md section 3.3; this requirement imports its claim D-36#7 as a firm rule, so changing it needs the author. Reason not stated in the decision. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The read API reports, per account, the date it is valued through and whether a revaluation is pending or running, and the UI shows it.

Source: OpenFief author decisions (docs/DESIGN.md section 3.3) - D-36#7
