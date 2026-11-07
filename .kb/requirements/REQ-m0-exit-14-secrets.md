---
title: Credentials must exist only as Kubernetes Secrets, mounted only into plugins granted them.
status: open
text_ref: design-decisions:D-200#14
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-14
  - lane:strict
semantic_text: Credentials must exist only as Kubernetes Secrets, mounted only into plugins granted them.
logic_claims:
  - CLAIM-55DBBD71E9A30914
semantic_clauses:
  - Credentials must exist only as Kubernetes Secrets, mounted only into plugins granted them.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 25e0fb3ac4204ad566bc240236f8eab234443e312981915ccf652ea77e2f0985
semantic_inventory:
  - claim_key: CLAIM-55DBBD71E9A30914
    claim_text: Credentials must exist only as Kubernetes Secrets, mounted only into plugins granted them
    role: normative
    status: modeled
    span:
      start: 0
      end: 89
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-200#14
  recorded_at: '2026-11-07T03:32:29.326Z'
id: REQ-m0-exit-14-secrets
type: req
---
Credentials must exist only as Kubernetes Secrets, mounted only into plugins granted them.

## Context

Author decision D-200: M0 is done when every one of its nineteen criteria holds. Each criterion is an end-to-end scenario on the firm requirement it checks, and the whole set passes on both kind and k3d (criterion 15); the scenarios that verify this criterion carry the tag acceptance:p10-14.

## Source

> **Secrets.** Credentials exist only as Kubernetes Secrets, mounted only into plugins granted them.

design-decisions:D-200#14 (docs/DESIGN.md, D-200, criterion 14)
