---
title: While a Plugin's SignatureVerified condition is True, the operator must record in its status the trust level, the kind and the name of the trusted signer, and must record none otherwise.
status: open
text_ref: docs/DESIGN.md#D-56
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-56
  - lane:strict
semantic_text: While a Plugin's SignatureVerified condition is True, the operator must record in its status the trust level, the kind and the name of the trusted signer, and must record none otherwise.
logic_claims:
  - CLAIM-FE4FB9C183DCA738
semantic_clauses:
  - While a Plugin's SignatureVerified condition is True, the operator must record in its status the trust level, the kind and the name of the trusted signer, and must record none otherwise.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d28b5b4e37b47216de9c87189300c79d9ae3ca525e2227229ebe9094af7f1ccf
semantic_inventory:
  - claim_key: CLAIM-FE4FB9C183DCA738
    claim_text: While a Plugin's SignatureVerified condition is True, the operator must record in its status the trust level, the kind and the name of the trusted signer, and must record none otherwise
    role: normative
    status: modeled
    span:
      start: 0
      end: 185
origin:
  kind: agent
  recorded_at: '2026-10-29T03:01:20.204Z'
id: REQ-operator-signature-signer-status
type: req
---
While a Plugin's SignatureVerified condition is True, the operator must record in its status the trust level, the kind and the name of the trusted signer, and must record none otherwise.

## Context

D-56 makes the UI show whether an image is verified or only signed; the task asks for signed by OpenFief's CI and signed by the owner as separate levels, which the condition's reason can't tell apart. An additive status field is the narrowest structured carrier. Modeled by Claude Code for the plugin status task; the author has not reviewed it yet.

## Source

> **[D-56]** The operator verifies every plugin image's signature itself before deploying it, so a `Plugin` applied through GitOps or plain `kubectl apply` is checked too. [...] It records in the `Plugin`'s status whether the image is verified (D-44) or only signed by a trusted signer, and the UI shows that status.

docs/DESIGN.md#D-56
