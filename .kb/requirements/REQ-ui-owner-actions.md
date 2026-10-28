---
title: The built-in UI must offer passkey management, script tokens and enrollment links to an owner-scoped session only, show a new script token's secret once, and show the core's reason when it refuses an action.
status: open
text_ref: docs/DESIGN.md#D-27
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-27
  - lane:strict
semantic_text: The built-in UI must offer passkey management, script tokens and enrollment links to an owner-scoped session only, show a new script token's secret once, and show the core's reason when it refuses an action.
logic_claims:
  - CLAIM-D912276E72D6F35F
semantic_clauses:
  - The built-in UI must offer passkey management, script tokens and enrollment links to an owner-scoped session only, show a new script token's secret once, and show the core's reason when it refuses an action.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a29498991db4cc246f64c20eb5bcf374e0667dcfb7781c329e892c3b43786452
semantic_inventory:
  - claim_key: CLAIM-D912276E72D6F35F
    claim_text: The built-in UI must offer passkey management, script tokens and enrollment links to an owner-scoped session only, show a new script token's secret once, and show the core's reason when it refuses an action
    role: normative
    status: modeled
    span:
      start: 0
      end: 206
origin:
  kind: agent
  recorded_at: '2026-10-28T19:49:51.748Z'
id: REQ-ui-owner-actions
type: req
---
The built-in UI must offer passkey management, script tokens and enrollment links to an owner-scoped session only, show a new script token's secret once, and show the core's reason when it refuses an action.

## Context

D-27 (script tokens), D-60 (enrollment links per hostname) and the author's decision that script tokens are created after a passkey login in the UI, from an owner session. The core's AuthService contract (proto/openfief/identity/v1/auth.proto) and the author's auth split (authenticators, then an opaque revocable session with a read or owner scope) given in chat to the orchestrator.

## Source

> Scripts get separate, scoped, expiring, read-only personal tokens.

docs/DESIGN.md#D-27
