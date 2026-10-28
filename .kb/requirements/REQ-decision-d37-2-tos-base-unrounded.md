---
title: The TOS yearly base must not be rounded.
status: open
text_ref: design-decisions:D-37#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-37-2
  - lane:strict
semantic_text: The TOS yearly base must not be rounded.
logic_claims:
  - CLAIM-031F45CBD456D547
semantic_clauses:
  - The TOS yearly base must not be rounded.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9ea5dd37154e72b12b5ae8f78ea0fb69ecc77143da802b999390af6dc1b78636
semantic_inventory:
  - claim_key: CLAIM-031F45CBD456D547
    claim_text: The TOS yearly base must not be rounded
    role: normative
    status: modeled
    span:
      start: 0
      end: 39
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-37#2
  recorded_at: '2026-10-28T06:30:45.834Z'
id: REQ-decision-d37-2-tos-base-unrounded
type: req
---
The TOS yearly base must not be rounded.

## Context

The author decided D-37 in docs/DESIGN.md section 3.3; this requirement imports its claim D-37#2 as a firm rule, so changing it needs the author. The TOS issue letter words the yearly base as rounded to grosze, but the issuer's published interest tables do not round it. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Where a treasury bond's issue letter and the issuer's published daily interest tables or worked examples disagree, OpenFief follows the tables and examples: the TOS yearly base is not rounded, and Belka tax is rounded half up to the grosz.

Source: OpenFief author decisions (docs/DESIGN.md section 3.3) - D-37#2
