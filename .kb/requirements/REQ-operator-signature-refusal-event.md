---
title: When a reconcile newly refuses a Plugin image signature, the operator must record one Warning event on the Plugin naming the refusal reason and message, and must not record it again while the refusal is unchanged.
status: open
text_ref: docs/DESIGN.md#D-56
tags:
  - strict-modeling
  - confidence:0.85
  - confidence-band:medium
  - provenance:docs-design-md-d-56
  - lane:strict
semantic_text: When a reconcile newly refuses a Plugin image signature, the operator must record one Warning event on the Plugin naming the refusal reason and message, and must not record it again while the refusal is unchanged.
logic_claims:
  - CLAIM-B31ADC2A8D96AE6D
semantic_clauses:
  - When a reconcile newly refuses a Plugin image signature, the operator must record one Warning event on the Plugin naming the refusal reason and message, and must not record it again while the refusal is unchanged.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a5c837df48cd4f54c321f77441274a12a58218c1280afba862c7568c61de20b5
semantic_inventory:
  - claim_key: CLAIM-B31ADC2A8D96AE6D
    claim_text: When a reconcile newly refuses a Plugin image signature, the operator must record one Warning event on the Plugin naming the refusal reason and message, and must not record it again while the refusal is unchanged
    role: condition
    status: modeled
    span:
      start: 0
      end: 212
origin:
  kind: agent
  recorded_at: '2026-11-01T13:51:18.226Z'
id: REQ-operator-signature-refusal-event
type: req
---
When a reconcile newly refuses a Plugin image signature, the operator must record one Warning event on the Plugin naming the refusal reason and message, and must not record it again while the refusal is unchanged.

## Context

D-56 has the operator check every plugin image itself and record the outcome in the Plugin's status. The orchestrating agent asked for the refusal of an unsigned or untrusted image to be visible in the Plugin's status and events (release readiness item R2), so an owner who applied a Plugin with kubectl sees why it does not run in kubectl describe. Recording only new refusals keeps a requeued or repeatedly reconciled Plugin from flooding its events. Modeled by Claude Code; the author has not reviewed it.

## Source

> the operator verifies every image itself ... and the status says Verified or SignedByTrustedSigner

docs/DESIGN.md#D-56
