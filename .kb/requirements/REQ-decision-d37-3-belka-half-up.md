---
title: Belka tax on treasury bonds must be rounded half up to the grosz.
status: open
text_ref: design-decisions:D-37#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-37-3
  - lane:strict
semantic_text: Belka tax on treasury bonds must be rounded half up to the grosz.
logic_claims:
  - CLAIM-CB3452FBA33E1DA7
semantic_clauses:
  - Belka tax on treasury bonds must be rounded half up to the grosz.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 87499f195b55c2aa9f1729f2d60bdcc1b9c5e7e334da88726a4636ae5121fbec
semantic_inventory:
  - claim_key: CLAIM-CB3452FBA33E1DA7
    claim_text: Belka tax on treasury bonds must be rounded half up to the grosz
    role: normative
    status: modeled
    span:
      start: 0
      end: 64
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-37#3
  recorded_at: '2026-10-28T06:30:55.102Z'
id: REQ-decision-d37-3-belka-half-up
type: req
---
Belka tax on treasury bonds must be rounded half up to the grosz.

## Context

The author decided D-37 in docs/DESIGN.md section 3.3; this requirement imports its claim D-37#3 as a firm rule, so changing it needs the author. The issuer's worked examples round Belka tax half up, which the author chose over other readings. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Where a treasury bond's issue letter and the issuer's published daily interest tables or worked examples disagree, OpenFief follows the tables and examples: the TOS yearly base is not rounded, and Belka tax is rounded half up to the grosz.

Source: OpenFief author decisions (docs/DESIGN.md section 3.3) - D-37#3
