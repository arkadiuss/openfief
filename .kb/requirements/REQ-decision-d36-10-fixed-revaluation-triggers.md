---
title: What triggers a revaluation must not be configurable.
status: open
text_ref: design-decisions:D-36#10
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-36-10
  - lane:strict
semantic_text: What triggers a revaluation must not be configurable.
logic_claims:
  - CLAIM-B80624281785542D
semantic_clauses:
  - What triggers a revaluation must not be configurable.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 24462355c40e79890a6804ba7f71446c385b26192204c4cef11e992b5545e1f3
semantic_inventory:
  - claim_key: CLAIM-B80624281785542D
    claim_text: What triggers a revaluation must not be configurable
    role: normative
    status: modeled
    span:
      start: 0
      end: 52
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-36#10
  recorded_at: '2026-10-28T06:30:28.886Z'
id: REQ-decision-d36-10-fixed-revaluation-triggers
type: req
---
What triggers a revaluation must not be configurable.

## Context

The author decided D-36 in docs/DESIGN.md section 3.3; this requirement imports its claim D-36#10 as a firm rule, so changing it needs the author. Reason not stated in the decision. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> What triggers a revaluation is not configurable.

Source: OpenFief author decisions (docs/DESIGN.md section 3.3) - D-36#10
