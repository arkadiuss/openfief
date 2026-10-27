---
title: The XTB connector must import closed CFD results as cash profit or loss.
status: open
text_ref: design-decisions:D-24#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-24-2
  - lane:strict
semantic_text: The XTB connector must import closed CFD results as cash profit or loss.
logic_claims:
  - CLAIM-3BE6EFE8C7367D4D
semantic_clauses:
  - The XTB connector must import closed CFD results as cash profit or loss.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8d50dfb52f7af264dcdf9449bc8f461d83192d5c7f8aa59e07cf77b340053452
semantic_inventory:
  - claim_key: CLAIM-3BE6EFE8C7367D4D
    claim_text: The XTB connector must import closed CFD results as cash profit or loss
    role: normative
    status: modeled
    span:
      start: 0
      end: 71
origin:
  kind: import
  ref: design-decisions:D-24#2
  recorded_at: '2026-10-27T20:12:40.785Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-A93F7CCF49981B85
type: req
---
The XTB connector must import closed CFD results as cash profit or loss.

## Source

> XTB in M0 covers real stocks and ETFs only. Closed CFD results are imported as cash profit or loss so cash reconciles; open CFD positions are flagged as unsupported.

Source: OpenFief author decisions (DESIGN.md section 10) - D-24#2
