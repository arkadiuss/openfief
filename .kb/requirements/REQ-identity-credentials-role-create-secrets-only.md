---
title: The Role of the Job may only create Secrets in the openfief namespace.
status: open
text_ref: design-decisions:D-216#5
tags:
  - strict-modeling
  - confidence:0.80
  - confidence-band:medium
  - lane:strict
  - provenance:design-decisions-d-216-5
semantic_text: The Role of the Job may only create Secrets in the openfief namespace.
logic_claims:
  - CLAIM-6CE2A76B13702F4B
semantic_clauses:
  - The Role of the Job may only create Secrets in the openfief namespace.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: dcdf3d162086bb5801ae6c892b2f13ae712a8c8b980371da47ac4f1c318c15e8
semantic_inventory:
  - claim_key: CLAIM-6CE2A76B13702F4B
    claim_text: The Role of the Job may only create Secrets in the openfief namespace
    role: normative
    status: modeled
    span:
      start: 0
      end: 69
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-216#5
  recorded_at: '2026-11-13T07:43:01.342Z'
id: REQ-identity-credentials-role-create-secrets-only
type: req
---
The Role of the Job may only create Secrets in the openfief namespace.

## Context

Part of D-216: the Job's ServiceAccount holds the right to create Secrets and no other right, not even to read one back (D-15). The release bundle's test fails unless that Role is the only one in the install that grants create on secrets. The decision states no further reason. The author recorded D-216 in DESIGN.md.

## Source

> The bootstrap install generates the platform credentials nothing else makes, such as SpiceDB's preshared key and its database role's password (D-164): a one-shot Job, run from the core's own signed image, creates each Secret only if it is missing and never reads, changes or rotates one, so an owner who feeds them from their own tooling first (D-29) keeps theirs. Its Role may only create Secrets in the `openfief` namespace.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-216#5