---
title: The XTB connector must flag open CFD positions as unsupported.
status: open
text_ref: design-decisions:D-24#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-24-3
  - lane:strict
semantic_text: The XTB connector must flag open CFD positions as unsupported.
logic_claims:
  - CLAIM-DD75D91D503761CF
semantic_clauses:
  - The XTB connector must flag open CFD positions as unsupported.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3684775998364276cec591a2fc24ded1275ccbf18f1e2cad6db602951a3669fd
semantic_inventory:
  - claim_key: CLAIM-DD75D91D503761CF
    claim_text: The XTB connector must flag open CFD positions as unsupported
    role: normative
    status: modeled
    span:
      start: 0
      end: 61
origin:
  kind: import
  ref: design-decisions:D-24#3
  recorded_at: '2026-10-27T20:12:41.205Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-A5EE8BC2AE68372A
type: req
---
The XTB connector must flag open CFD positions as unsupported.

## Source

> XTB in M0 covers real stocks and ETFs only. Closed CFD results are imported as cash profit or loss so cash reconciles; open CFD positions are flagged as unsupported.

Source: OpenFief author decisions (DESIGN.md section 10) - D-24#3
