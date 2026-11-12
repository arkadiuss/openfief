---
title: A Plugin without a matching PluginGrant must not be deployed, an upgrade requesting a new egress host must stay on the old version until granted, an image referenced by tag or without a valid cosign signature from a trusted signer must never be deployed and the Plugin's status must say why, including when applied with plain kubectl (D-56), and the Plugin's status must show verified only for an image the catalog pins.
status: open
tags:
  - strict-modeling
  - confidence:0.80
  - confidence-band:medium
  - provenance:design-decisions-d-200-11
  - lane:strict
text_ref: design-decisions:D-200#11
origin:
  kind: import
  ref: design-decisions:D-200#11
  recorded_at: '2026-11-12T19:56:11.048Z'
semantic_text: A Plugin without a matching PluginGrant must not be deployed, an upgrade requesting a new egress host must stay on the old version until granted, an image referenced by tag or without a valid cosign signature from a trusted signer must never be deployed and the Plugin's status must say why, including when applied with plain kubectl (D-56), and the Plugin's status must show verified only for an image the catalog pins.
logic_claims:
  - CLAIM-53A1FBE82483C248
semantic_clauses:
  - A Plugin without a matching PluginGrant must not be deployed, an upgrade requesting a new egress host must stay on the old version until granted, an image referenced by tag or without a valid cosign signature from a trusted signer must never be deployed and the Plugin's status must say why, including when applied with plain kubectl (D-56), and the Plugin's status must show verified only for an image the catalog pins.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: dfbbda892d98ee3611bba6743a4be660ded89bcbad74dfb59d9a8e216bb681ac
semantic_inventory:
  - claim_key: CLAIM-53A1FBE82483C248
    claim_text: A Plugin without a matching PluginGrant must not be deployed, an upgrade requesting a new egress host must stay on the old version until granted, an image referenced by tag or without a valid cosign signature from a trusted signer must never be deployed and the Plugin's status must say why, including when applied with plain kubectl (D-56), and the Plugin's status must show verified only for an image the catalog pins
    role: normative
    status: modeled
    span:
      start: 0
      end: 419
id: REQ-m0-exit-11-operator-v2
type: req
---
A Plugin without a matching PluginGrant must not be deployed, an upgrade requesting a new egress host must stay on the old version until granted, an image referenced by tag or without a valid cosign signature from a trusted signer must never be deployed and the Plugin's status must say why, including when applied with plain kubectl (D-56), and the Plugin's status must show verified only for an image the catalog pins.

## Context

Author decision D-200: M0 is done when every one of its nineteen criteria holds, each an end-to-end scenario on the firm requirement it checks. The author reworded criterion 11 (commit a8ed88f) so that rejected means what D-56 makes it mean: the operator verifies signatures itself, so an image refused is never deployed and the Plugin's status says why, also for a Plugin applied with plain kubectl. The scenarios that verify this criterion carry the tag acceptance:p10-11. This requirement replaces REQ-m0-exit-11-operator, which kept the earlier wording.

## Source

> **Operator.** A `Plugin` without a matching `PluginGrant` is not deployed. An upgrade requesting a new egress host stays on the old version until granted. An image referenced by tag, or without a valid cosign signature from a trusted signer, is never deployed, and the `Plugin`'s status says why, including when applied with plain `kubectl` (D-56). The `Plugin`'s status shows verified only for an image the catalog pins.

design-decisions:D-200#11 (docs/DESIGN.md, D-200, criterion 11)

## Why this changed

D-200 #11 was reworded by the author to D-56's meaning of rejected, so the recorded statement no longer matched the decision it imports.