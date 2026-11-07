---
title: Uninstalling a connector instance must remove its pod, grant, secrets and allow-list, while its imported transactions remain, marked as coming from a removed instance.
status: open
text_ref: design-decisions:D-200#18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-18
  - lane:strict
semantic_text: Uninstalling a connector instance must remove its pod, grant, secrets and allow-list, while its imported transactions remain, marked as coming from a removed instance.
logic_claims:
  - CLAIM-E3DC34EF944717D0
semantic_clauses:
  - Uninstalling a connector instance must remove its pod, grant, secrets and allow-list, while its imported transactions remain, marked as coming from a removed instance.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 95d76b69051abe00a58e4af2d195bc66b2a990938fef2b59417af9f4f1474604
semantic_inventory:
  - claim_key: CLAIM-E3DC34EF944717D0
    claim_text: Uninstalling a connector instance must remove its pod, grant, secrets and allow-list, while its imported transactions remain, marked as coming from a removed instance
    role: normative
    status: modeled
    span:
      start: 0
      end: 166
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-200#18
  recorded_at: '2026-11-07T03:32:50.253Z'
id: REQ-m0-exit-18-uninstall
type: req
---
Uninstalling a connector instance must remove its pod, grant, secrets and allow-list, while its imported transactions remain, marked as coming from a removed instance.

## Context

Author decision D-200: M0 is done when every one of its nineteen criteria holds. Each criterion is an end-to-end scenario on the firm requirement it checks, and the whole set passes on both kind and k3d (criterion 15); the scenarios that verify this criterion carry the tag acceptance:p10-18.

## Source

> **Uninstall.** Uninstalling a connector instance removes its pod, grant, secrets and allow-list; its imported transactions remain, marked as coming from a removed instance.

design-decisions:D-200#18 (docs/DESIGN.md, D-200, criterion 18)
