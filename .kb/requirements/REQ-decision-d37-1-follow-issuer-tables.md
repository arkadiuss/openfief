---
title: Where a treasury bond's issue letter disagrees with the issuer's published daily interest tables or worked examples, bond valuation must follow the tables and examples.
status: open
text_ref: design-decisions:D-37#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-37-1
  - lane:strict
semantic_text: Where a treasury bond's issue letter disagrees with the issuer's published daily interest tables or worked examples, bond valuation must follow the tables and examples.
logic_claims:
  - CLAIM-93F09732C0EACFD9
semantic_clauses:
  - Where a treasury bond's issue letter disagrees with the issuer's published daily interest tables or worked examples, bond valuation must follow the tables and examples.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 19f4926ef27fceed799628f70c16eaf4859cb361ffad09e481422dcfe9bc8e91
semantic_inventory:
  - claim_key: CLAIM-93F09732C0EACFD9
    claim_text: Where a treasury bond's issue letter disagrees with the issuer's published daily interest tables or worked examples, bond valuation must follow the tables and examples
    role: normative
    status: modeled
    span:
      start: 0
      end: 167
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-37#1
  recorded_at: '2026-10-28T06:30:36.452Z'
id: REQ-decision-d37-1-follow-issuer-tables
type: req
---
Where a treasury bond's issue letter disagrees with the issuer's published daily interest tables or worked examples, bond valuation must follow the tables and examples.

## Context

The author decided D-37 in docs/DESIGN.md section 3.3; this requirement imports its claim D-37#1 as a firm rule, so changing it needs the author. The issue letters and the issuer's published tables and examples disagree in places; the author chose the tables and examples. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Where a treasury bond's issue letter and the issuer's published daily interest tables or worked examples disagree, OpenFief follows the tables and examples: the TOS yearly base is not rounded, and Belka tax is rounded half up to the grosz.

Source: OpenFief author decisions (docs/DESIGN.md section 3.3) - D-37#1
