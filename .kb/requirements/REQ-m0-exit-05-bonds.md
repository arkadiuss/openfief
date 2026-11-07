---
title: For every held bond series the gross accrued value must match an open-source reference valuer on the same date to within 0.01 PLN, or, where no such valuer can be pinned, a hand calculation from the Ministry of Finance's issue letters, one per bond type and held series, with its formulas recorded in research; net-if-redeemed must match a hand calculation for at least one series of each type held; and totals must use the configured basis, net by default.
status: open
text_ref: design-decisions:D-200#5
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-5
  - lane:strict
semantic_text: For every held bond series the gross accrued value must match an open-source reference valuer on the same date to within 0.01 PLN, or, where no such valuer can be pinned, a hand calculation from the Ministry of Finance's issue letters, one per bond type and held series, with its formulas recorded in research; net-if-redeemed must match a hand calculation for at least one series of each type held; and totals must use the configured basis, net by default.
logic_claims:
  - CLAIM-FDEACBB00F55AC31
semantic_clauses:
  - For every held bond series the gross accrued value must match an open-source reference valuer on the same date to within 0.01 PLN, or, where no such valuer can be pinned, a hand calculation from the Ministry of Finance's issue letters, one per bond type and held series, with its formulas recorded in research; net-if-redeemed must match a hand calculation for at least one series of each type held; and totals must use the configured basis, net by default.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5eb1f57bd20c09fdd0f802779947146d6aaf95f11553d151314fd71cd9ef6327
semantic_inventory:
  - claim_key: CLAIM-FDEACBB00F55AC31
    claim_text: For every held bond series the gross accrued value must match an open-source reference valuer on the same date to within 0.01 PLN, or, where no such valuer can be pinned, a hand calculation from the Ministry of Finance's issue letters, one per bond type and held series, with its formulas recorded in research; net-if-redeemed must match a hand calculation for at least one series of each type held; and totals must use the configured basis, net by default
    role: normative
    status: modeled
    span:
      start: 0
      end: 456
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-200#5
  recorded_at: '2026-11-07T03:31:45.605Z'
id: REQ-m0-exit-05-bonds
type: req
---
For every held bond series the gross accrued value must match an open-source reference valuer on the same date to within 0.01 PLN, or, where no such valuer can be pinned, a hand calculation from the Ministry of Finance's issue letters, one per bond type and held series, with its formulas recorded in research; net-if-redeemed must match a hand calculation for at least one series of each type held; and totals must use the configured basis, net by default.

## Context

Author decision D-200: M0 is done when every one of its nineteen criteria holds. Each criterion is an end-to-end scenario on the firm requirement it checks, and the whole set passes on both kind and k3d (criterion 15); the scenarios that verify this criterion carry the tag acceptance:p10-5.

## Source

> **Bonds.** For every held series, the gross accrued value matches an open-source reference valuer on the same date to within 0.01 PLN, or, where no such valuer can be pinned, a hand calculation from the Ministry of Finance's issue letters, one per bond type and per held series, with its formulas recorded in research, and net-if-redeemed matches a hand calculation for at least one series of each type held. Totals use the configured basis, net by default.

design-decisions:D-200#5 (docs/DESIGN.md, D-200, criterion 5)
