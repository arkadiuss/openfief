---
title: The operator's ServiceAccount must not be able to create or change anything outside the openfief namespace.
status: open
text_ref: design-decisions:D-200#12
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-12
  - lane:strict
semantic_text: The operator's ServiceAccount must not be able to create or change anything outside the openfief namespace.
logic_claims:
  - CLAIM-4BE5EDCEFAA75586
semantic_clauses:
  - The operator's ServiceAccount must not be able to create or change anything outside the openfief namespace.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 52e369c0eab54df97af4d88abf6b669b8c4dd731024e58b36bf2742ef6bb847a
semantic_inventory:
  - claim_key: CLAIM-4BE5EDCEFAA75586
    claim_text: The operator's ServiceAccount must not be able to create or change anything outside the openfief namespace
    role: normative
    status: modeled
    span:
      start: 0
      end: 106
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-200#12
  recorded_at: '2026-11-07T03:32:19.747Z'
id: REQ-m0-exit-12-operator-scope
type: req
---
The operator's ServiceAccount must not be able to create or change anything outside the openfief namespace.

## Context

Author decision D-200: M0 is done when every one of its nineteen criteria holds. Each criterion is an end-to-end scenario on the firm requirement it checks, and the whole set passes on both kind and k3d (criterion 15); the scenarios that verify this criterion carry the tag acceptance:p10-12.

## Source

> **Operator scope.** The operator's ServiceAccount cannot create or change anything outside the `openfief` namespace.

design-decisions:D-200#12 (docs/DESIGN.md, D-200, criterion 12)
