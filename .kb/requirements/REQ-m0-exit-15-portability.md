---
title: The whole M0 acceptance suite must pass on both kind and k3d.
status: open
text_ref: design-decisions:D-200#15
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-15
  - lane:strict
semantic_text: The whole M0 acceptance suite must pass on both kind and k3d.
logic_claims:
  - CLAIM-3D494677367C017F
semantic_clauses:
  - The whole M0 acceptance suite must pass on both kind and k3d.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0a9d3fdb10619e95dbc32eae0243ad4c27361736d67e3ac94c15d7a0b8353553
semantic_inventory:
  - claim_key: CLAIM-3D494677367C017F
    claim_text: The whole M0 acceptance suite must pass on both kind and k3d
    role: normative
    status: modeled
    span:
      start: 0
      end: 60
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-200#15
  recorded_at: '2026-11-07T03:32:34.200Z'
id: REQ-m0-exit-15-portability
type: req
---
The whole M0 acceptance suite must pass on both kind and k3d.

## Context

Author decision D-200: M0 is done when every one of its nineteen criteria holds. Each criterion is an end-to-end scenario on the firm requirement it checks, and the whole set passes on both kind and k3d (criterion 15); the scenarios that verify this criterion carry the tag acceptance:p10-15. CI's e2e job runs the suite on a k3d and on a kind cluster; this holds for the scenarios the suite verifies today and grows with them.

## Source

> **Portability.** The whole acceptance suite passes on both kind and k3d.

design-decisions:D-200#15 (docs/DESIGN.md, D-200, criterion 15)
