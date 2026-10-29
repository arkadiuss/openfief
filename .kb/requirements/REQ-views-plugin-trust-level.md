---
title: 'The core must derive one trust level per installed plugin from its status: verified, signed by OpenFief''s CI, by a catalog publisher, by the owner or by a trusted signer, refused, or not checked.'
status: open
text_ref: docs/DESIGN.md#D-56
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-56
  - lane:strict
semantic_text: 'The core must derive one trust level per installed plugin from its status: verified, signed by OpenFief''s CI, by a catalog publisher, by the owner or by a trusted signer, refused, or not checked.'
logic_claims:
  - CLAIM-4582D311FFD68BF2
semantic_clauses:
  - 'The core must derive one trust level per installed plugin from its status: verified, signed by OpenFief''s CI, by a catalog publisher, by the owner or by a trusted signer, refused, or not checked.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 42fe766a527818db9f060b97418bfbf468eb49103cc66fcc4caeee88cd41bb38
semantic_inventory:
  - claim_key: CLAIM-4582D311FFD68BF2
    claim_text: 'The core must derive one trust level per installed plugin from its status: verified, signed by OpenFief''s CI, by a catalog publisher, by the owner or by a trusted signer, refused, or not checked'
    role: normative
    status: modeled
    span:
      start: 0
      end: 194
origin:
  kind: agent
  recorded_at: '2026-10-29T03:29:39.319Z'
id: REQ-views-plugin-trust-level
type: req
---
The core must derive one trust level per installed plugin from its status: verified, signed by OpenFief's CI, by a catalog publisher, by the owner or by a trusted signer, refused, or not checked.

## Context

D-56: the UI shows whether an image is verified (D-44) or only signed by a trusted signer; D-45 names the three kinds of trusted signer. Showing the signer kind, refused and not checked are this task's reading. Modeled by Claude Code for the plugin status task; the author has not reviewed it yet.

## Source

> **[D-56]** The operator verifies every plugin image's signature itself before deploying it, so a `Plugin` applied through GitOps or plain `kubectl apply` is checked too. [...] It records in the `Plugin`'s status whether the image is verified (D-44) or only signed by a trusted signer, and the UI shows that status.

docs/DESIGN.md#D-56
