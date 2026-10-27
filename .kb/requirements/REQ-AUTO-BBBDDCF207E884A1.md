---
title: The operator must hold permissions only within the openfief namespace.
status: open
text_ref: design-decisions:D-15#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-15-1
  - lane:strict
semantic_text: The operator must hold permissions only within the openfief namespace.
logic_claims:
  - CLAIM-161F13F619C48276
semantic_clauses:
  - The operator must hold permissions only within the openfief namespace.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ba107e7d03d4807c656062ccb2c905e38481ad2accf0d5373645958faf52cb4e
semantic_inventory:
  - claim_key: CLAIM-161F13F619C48276
    claim_text: The operator must hold permissions only within the openfief namespace
    role: normative
    status: modeled
    span:
      start: 0
      end: 69
origin:
  kind: import
  ref: design-decisions:D-15#1
  recorded_at: '2026-10-27T20:12:29.698Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-BBBDDCF207E884A1
type: req
---
The operator must hold permissions only within the openfief namespace.

## Source

> The operator is namespace-scoped: a Role in the `openfief` namespace only. The CRDs are installed once by a cluster admin. Namespace-per-agent is revisited at M2.

Source: OpenFief author decisions (DESIGN.md section 10) - D-15#1
