---
title: The XTB connector must import only real stocks and ETFs as holdings in M0.
status: open
text_ref: design-decisions:D-24#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-24-1
  - lane:strict
semantic_text: The XTB connector must import only real stocks and ETFs as holdings in M0.
logic_claims:
  - CLAIM-F0ECDF543C6FE213
semantic_clauses:
  - The XTB connector must import only real stocks and ETFs as holdings in M0.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 35291187bb271f3c3b0204e2c956f19bc06af82fe658ca7bf466db5444953450
semantic_inventory:
  - claim_key: CLAIM-F0ECDF543C6FE213
    claim_text: The XTB connector must import only real stocks and ETFs as holdings in M0
    role: normative
    status: modeled
    span:
      start: 0
      end: 73
origin:
  kind: import
  ref: design-decisions:D-24#1
  recorded_at: '2026-10-27T20:12:40.404Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-39B3A0CB06192CFE
type: req
---
The XTB connector must import only real stocks and ETFs as holdings in M0.

## Source

> XTB in M0 covers real stocks and ETFs only. Closed CFD results are imported as cash profit or loss so cash reconciles; open CFD positions are flagged as unsupported.

Source: OpenFief author decisions (DESIGN.md section 10) - D-24#1
