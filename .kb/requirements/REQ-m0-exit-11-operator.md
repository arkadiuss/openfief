---
title: A Plugin without a matching PluginGrant must not be deployed, an upgrade requesting a new egress host must stay on the old version until granted, an image referenced by tag or without a valid cosign signature from a trusted signer must be rejected, including when applied with plain kubectl, and the Plugin's status must show verified only for an image the catalog pins.
status: closed
text_ref: design-decisions:D-200#11
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-11
  - lane:strict
semantic_text: A Plugin without a matching PluginGrant must not be deployed, an upgrade requesting a new egress host must stay on the old version until granted, an image referenced by tag or without a valid cosign signature from a trusted signer must be rejected, including when applied with plain kubectl, and the Plugin's status must show verified only for an image the catalog pins.
logic_claims:
  - CLAIM-5B6D235B2E3277CA
semantic_clauses:
  - A Plugin without a matching PluginGrant must not be deployed, an upgrade requesting a new egress host must stay on the old version until granted, an image referenced by tag or without a valid cosign signature from a trusted signer must be rejected, including when applied with plain kubectl, and the Plugin's status must show verified only for an image the catalog pins.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d67fbf008a3451cdfd6f088784a4e15fc5a48a48288c3e393a8fdda667d23b84
semantic_inventory:
  - claim_key: CLAIM-5B6D235B2E3277CA
    claim_text: A Plugin without a matching PluginGrant must not be deployed, an upgrade requesting a new egress host must stay on the old version until granted, an image referenced by tag or without a valid cosign signature from a trusted signer must be rejected, including when applied with plain kubectl, and the Plugin's status must show verified only for an image the catalog pins
    role: normative
    status: modeled
    span:
      start: 0
      end: 369
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-200#11
  recorded_at: '2026-11-07T03:32:14.909Z'
id: REQ-m0-exit-11-operator
type: req
---
A Plugin without a matching PluginGrant must not be deployed, an upgrade requesting a new egress host must stay on the old version until granted, an image referenced by tag or without a valid cosign signature from a trusted signer must be rejected, including when applied with plain kubectl, and the Plugin's status must show verified only for an image the catalog pins.

## Context

Author decision D-200: M0 is done when every one of its nineteen criteria holds. Each criterion is an end-to-end scenario on the firm requirement it checks, and the whole set passes on both kind and k3d (criterion 15); the scenarios that verify this criterion carry the tag acceptance:p10-11.

## Source

> **Operator.** A `Plugin` without a matching `PluginGrant` is not deployed. An upgrade requesting a new egress host stays on the old version until granted. An image referenced by tag, or without a valid cosign signature from a trusted signer, is rejected, including when applied with plain `kubectl`. The `Plugin`'s status shows verified only for an image the catalog pins.

design-decisions:D-200#11 (docs/DESIGN.md, D-200, criterion 11)
